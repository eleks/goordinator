package common

import (
  "encoding/gob"
  "encoding/binary"
  "io"
)

type Tasker interface {
  gob.GobEncoder
  gob.GobDecoder
  GrindIntoSubtasks(n uint32) ([]Tasker, error)
  GetID() uint32
  GetSubTask(i uint32, gN uint32) Tasker
  GetSize() uint32
  Dump(w io.Writer) error
}

func WriteTaskers(w io.Writer, params []Tasker) (err error) {
  p := uint32(len(params))
  err = binary.Write(w, binary.BigEndian, p)
  if err != nil { return err }

  var buf []byte

  for _, t := range params {
    buf, err = t.GobEncode()
    if err == nil {
      gd := GenericData{uint32(len(buf)), buf}
      err = gd.Write(w)

      if err != nil {
        return err
      }
    }
  }

  return nil
}
