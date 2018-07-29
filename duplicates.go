package main

import (
	"crypto/md5"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"sync"
	"sync/atomic"
)

// WalkedFile a type of struct
type WalkedFile struct {
	path string
	file os.FileInfo
}

var (
	singleThread  = false
	delete        = false
	visitCount    int64
	previous      = ""
	fileCount     int64
	dupCount      int64
	minSize       int64
	filenameMatch = "*"
	filenameRegex *regexp.Regexp
	duplicates    = struct {
		sync.RWMutex
		m map[string][]string
	}{m: make(map[string][]string)}
	noStats      bool
	walkProgress *Progress
	walkFiles    []*WalkedFile
)

func scanAndHashFile(path string, f os.FileInfo, progress *Progress) {
	if !f.IsDir() && f.Size() > minSize && (filenameMatch == "*" || filenameRegex.MatchString(f.Name())) {
		atomic.AddInt64(&fileCount, 1)
		file, err := os.Open(path)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%s\n", err.Error())
		} else {
			md5 := md5.New()
			io.Copy(md5, file)
			var hash = fmt.Sprintf("%x", md5.Sum(nil))
			//fmt.Printf("%s\t%s\t%d bytes\n", path, hash, f.Size())
			file.Close()
			duplicates.Lock()
			duplicates.m[hash] = append(duplicates.m[hash], path)
			duplicates.Unlock()
			progress.increment()
		}
	}
}

func worker(workerID int, jobs <-chan *WalkedFile, results chan<- int, progress *Progress) {
	for file := range jobs {
		//fmt.Println("hashing ", file.path, " on worker ", workerId)
		scanAndHashFile(file.path, file.file, progress)
		results <- 0
	}
}

func computeHashes() {
	walkProgress := creatProgress("Scanning %d files ...", &noStats)
	jobs := make(chan *WalkedFile, visitCount)
	results := make(chan int, visitCount)
	if singleThread {
		go worker(1, jobs, results, walkProgress)
	} else {
		for w := 1; w <= runtime.NumCPU(); w++ {
			go worker(w, jobs, results, walkProgress)
		}
	}
	for _, file := range walkFiles {
		jobs <- file
	}
	close(jobs)
	for _ = range walkFiles {
		<-results
	}
	walkProgress.delete()
}

func visitFile(path string, f os.FileInfo, err error) error {
	visitCount++
	if !f.IsDir() && f.Size() > minSize && (filenameMatch == "*" || filenameRegex.MatchString(f.Name())) {
		walkFiles = append(walkFiles, &WalkedFile{path: path, file: f})
		walkProgress.increment()
	}
	return nil
}

func deleteFile(path string) {
	fmt.Println("Deleting " + path)
	err := os.Remove(path)
	if err != nil {
		fmt.Printf("Error deleting file: %s \n", path)
	}
}

func main() {
	flag.Int64Var(&minSize, "size", 1, "Minimum size in bytes for a file")
	flag.StringVar(&filenameMatch, "name", "*", "Filename pattern")
	flag.BoolVar(&noStats, "nostats", false, "Do no output stats")
	flag.BoolVar(&singleThread, "single", false, "Work on only one thread")
	flag.BoolVar(&delete, "delete", true, "Delete duplicate files")
	var help = flag.Bool("h", false, "Display this message")
	flag.Parse()
	if *help {
		fmt.Println("duplicates is a command line tool to find duplicate files in a folder")
		fmt.Println("usage: duplicates [options...] path")
		flag.PrintDefaults()
		os.Exit(0)
	}
	if len(flag.Args()) < 1 {
		fmt.Fprintf(os.Stderr, "You have to specify at least a directory to explore ...\n")
		os.Exit(-1)
	}
	root := flag.Arg(0)
	walkProgress = creatProgress("Walking through %d files ...", &noStats)
	if !noStats {
		fmt.Printf("\nSearching duplicates in '%s' with name that match '%s' and minimum size '%d' bytes\n\n", root, filenameMatch, minSize)
	}
	r, _ := regexp.Compile(filenameMatch)
	filenameRegex = r
	filepath.Walk(root, visitFile)
	walkProgress.delete()
	computeHashes()
	for _, v := range duplicates.m {
		if len(v) > 1 {
			dupCount++
		}
	}
	fmt.Printf("/n /n /n")
	for _, v := range duplicates.m {
		if len(v) > 1 {
			for i, file := range v {
				fmt.Printf("%s\n", file)
				if i > 0 && delete {
					deleteFile(file)
				}
			}
			fmt.Println("---------")
		}
	}
	if !noStats {
		fmt.Printf("\nFound %d duplicates from %d files in %s with options { size: '%d', name: '%s' }\n", dupCount, fileCount, root, minSize, filenameMatch)
	}
	os.Exit(0)
}
