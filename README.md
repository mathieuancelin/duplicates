# duplicates

File duplicates finder

Sometimes you need to find duplicates files on your disk. You can use this tool to do it. It uses a MD5 hash to identify duplicate files. You can also use some options to filter files by names an minimum size (in bytes).

## usage

```
usage: duplicates [options...] path

  -h          Display the help message
  -name       Filename pattern
  -nostats    Do no output stats
  -single     Work in single threaded mode
  -size       Minimum size in bytes for a file
  -delete     Deletes duplicate files
```

## examples

```
$ duplicates /tmp
$ duplicates -name .mp3 /tmp
$ duplicates -size 2056 /tmp
$ duplicates -size 2056 -name .mp3 /tmp
$ duplicates -nostats -size 2056 -name .mp3 /tmp > duplicates.txt
```

## install

- from source

```
go get github.com/mathieuancelin/duplicates
```

- binaries

[![Gobuild Download](http://gobuild.io/badge/github.com/mathieuancelin/duplicates/downloads.svg)](http://gobuild.io/github.com/mathieuancelin/duplicates)
