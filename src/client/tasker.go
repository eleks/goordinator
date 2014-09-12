package main

import (
  "../common"
  "encoding/gob"
  "io"
)

type Tasker interface {
  gob.GobEncoder
  gob.GobDecoder
  GrindIntoSubtasks(n int) ([]Tasker, error)
  MergeSubtasks(subtasks []Tasker) error
  GetID() int
  GetSubTask(i int, gN int) Tasker
  GetSize() uint32
  Dump(w io.Writer) error
}
