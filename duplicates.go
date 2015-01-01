package main

import (
  "path/filepath"
  "crypto/md5"
  "os"
  "flag"
  "fmt"
  "io"
  "regexp"
  "runtime"
)

// ==================================================================================================

type Progress struct {
  notdisplay *bool
  pattern string
  previous string
  count int64  
}

func (pg *Progress) delete() {
  if (!*pg.notdisplay) {
    for j := 0; j <= len(pg.previous); j++ {
      fmt.Print("\b")
    }
  }
}

func (pg *Progress) displayToConsole() {
  if (!*pg.notdisplay) {
    pg.previous = fmt.Sprintf(pg.pattern, pg.count)
    fmt.Print(pg.previous)
  }
}

func (pg *Progress) increment() {
  pg.count++
  if (!*pg.notdisplay) {
    pg.delete()
    pg.displayToConsole()
  }
}

func creatProgress(pattern string, notdisplay *bool) (pg *Progress) {
  pg = &Progress{
    notdisplay: notdisplay,
    pattern: pattern,
    previous: "",
    count:   0,
  }
  return pg
}

// ==================================================================================================

type WalkedFile struct {
  path string
  file os.FileInfo 
}

var (
  singleThread bool = false
  visitCount int64 = 0
  previous string = ""
  fileCount int64 = 0
  dupCount int64 = 0
  minSize int64 = 0
  filenameMatch string = "*"
  filenameRegex *regexp.Regexp
  duplicates map[string][]string = make(map[string][]string)
  noStats bool
  walkProgress *Progress
  walkFiles []*WalkedFile
)

func scanAndHashFile(path string, f os.FileInfo, progress *Progress) {
  if (!f.IsDir() && f.Size() > minSize && (filenameMatch == "*" || filenameRegex.MatchString(f.Name()))) {
    fileCount++
    file, err := os.Open(path)
    if err != nil {
      fmt.Fprintf(os.Stderr, "%s\n", err.Error())
    } else {
      md5 := md5.New()
      io.Copy(md5, file)
      var hash string = fmt.Sprintf("%x", md5.Sum(nil))
      //fmt.Printf("%s\t%s\t%d bytes\n", path, hash, f.Size()) 
      file.Close()
      duplicates[hash] = append(duplicates[hash], path)
      progress.increment()
    }
  }  
}

func worker(workerId int, jobs <-chan *WalkedFile, results chan<- int, progress *Progress) {
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
  if (!f.IsDir() && f.Size() > minSize && (filenameMatch == "*" || filenameRegex.MatchString(f.Name()))) {
    walkFiles = append(walkFiles, &WalkedFile { path: path, file: f, })
    walkProgress.increment()
  }
  return nil  
}

func main() {
  flag.Int64Var(&minSize, "size", 1, "Minimum size in bytes for a file")
  flag.StringVar(&filenameMatch, "name", "*", "Filename pattern")
  flag.BoolVar(&noStats, "nostats", false, "Do no output stats")
  flag.BoolVar(&singleThread, "single", false, "Work on only one thread")
  var help = flag.Bool("h", false, "Display this message")
  flag.Parse()
  if (*help) {
    fmt.Println("\nduplicates is a command line tool to find duplicate files in a folder\n")
    fmt.Println("usage: duplicates [options...] path\n")
    flag.PrintDefaults()
    os.Exit(0)
  }
  if (len(flag.Args()) < 1) {
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
  for _, v := range duplicates {
    if (len(v) > 1) {
      dupCount++ 
    }
  }
  if !noStats {
    fmt.Printf("Found %d duplicates from %d files in %s with options { size: '%d', name: '%s' }\n\n", dupCount, fileCount, root, minSize, filenameMatch)
  }
  for _, v := range duplicates {
    if (len(v) > 1) {
      for _, file := range v {
        fmt.Printf("%s\n", file)
      }
      fmt.Printf("\n")
    }
  }
  if !noStats {
    fmt.Printf("\nFound %d duplicates from %d files in %s with options { size: '%d', name: '%s' }\n", dupCount, fileCount, root, minSize, filenameMatch)
  }
  os.Exit(0)
}