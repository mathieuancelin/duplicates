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

// TODO : scan files on multiple threads
// TODO : more options
// TODO : parse sizes

var (
  fileCount int64 = 0
  dupCount int64 = 0
  minSize int64 = 0
  filenameMatch string = "*"
  filenameRegex *regexp.Regexp
  duplicates map[string][]string = make(map[string][]string)
  noStats bool
)

func walkpath(path string, f os.FileInfo, err error) error {
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
    }
  }
  return nil  
}

func main() {
  flag.Int64Var(&minSize, "size", 1, "Minimum size in bytes for a file")
  flag.StringVar(&filenameMatch, "name", "*", "Filename pattern")
  flag.BoolVar(&noStats, "nostats", false, "Do no output stats")
  flag.Parse()
  root := flag.Arg(0) 
  if !noStats {
    fmt.Printf("\nSearching duplicates in '%s' with name that match '%s' and minimum size '%d' bytes\n\n", root, filenameMatch, minSize)
  }
  r, _ := regexp.Compile(filenameMatch)
  filenameRegex = r
  filepath.Walk(root, walkpath)
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