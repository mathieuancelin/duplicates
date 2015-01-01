package main

import (
  "fmt"
)

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