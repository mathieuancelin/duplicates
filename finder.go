package main

import (
  "path/filepath"
  "crypto/md5"
  "os"
  "flag"
  "fmt"
  "io"
  "regexp"
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

// TODO : scan files on multiple threads
// TODO : more options
// TODO : parse sizes

var (
  previous string = ""
  fileCount int64 = 0
  dupCount int64 = 0
  minSize int64 = 0
  filenameMatch string = "*"
  filenameRegex *regexp.Regexp
  duplicates map[string][]string = make(map[string][]string)
  noStats bool
  progress *Progress
)

func visitFile(path string, f os.FileInfo, err error) error {
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
  return nil  
}

func main() {
  flag.Int64Var(&minSize, "size", 1, "Minimum size in bytes for a file")
  flag.StringVar(&filenameMatch, "name", "*", "Filename pattern")
  flag.BoolVar(&noStats, "nostats", false, "Do no output stats")
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
  progress = creatProgress("Scanning %d files ...", &noStats)
  if !noStats {
    fmt.Printf("\nSearching duplicates in '%s' with name that match '%s' and minimum size '%d' bytes\n\n", root, filenameMatch, minSize)
  }
  r, _ := regexp.Compile(filenameMatch)
  filenameRegex = r
  filepath.Walk(root, visitFile)
  progress.delete()
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