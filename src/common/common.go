package common

import (
  "net"
)

type WorkerStatus byte
const (
  WReady WorkerStatus = iota
  WIdle
  WBusyAvailable
  WBusyFull
  WFinalized
)

type WorkerOperation byte
const (
  WInit WorkerOperation = iota
  WHealthCheck
  WGetTask
  WTaskCompeted
  WSendResult
)

type Task struct {
     id int
  // starting index and size of inner
  // multidimensional array of data
  index []int
   size []int
}

type Socket struct {
  Conn net.Conn
  Done chan bool
}

func (s Socket) Read(b []byte) (int, error) { return s.Conn.Read(b) }
func (s Socket) Write(b []byte) (int, error) { return s.Conn.Write(b) }

func (s Socket) Close() error {
  s.Done <- true
  return nil
}

type ClientStatus byte
const (
  CIdle ClientStatus = iota
  CBusy
)

type ClientOperation byte
const (
  CInitSession ClientOperation = iota
  CInputParameters
  CRunComputation
  CGetResult
)

type Parameter struct {
  Size int
  Data []byte
}
