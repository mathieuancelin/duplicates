# duplicates

File duplicates finder

Sometimes you need to find duplicates files on your disk. You can use this tool to do it. It uses a MD5 hash to identify duplicate files. You can also use some options to filter files by names an minimum size (in bytes).

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
