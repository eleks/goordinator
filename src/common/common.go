package common

import (
  "net"
  "io"
  "encoding/binary"  
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
  ID int
  Parameters InputParameters
}

type Socket struct {
  Conn net.Conn
  Done chan bool
}

func (s Socket) RemoteAddr() net.Addr { return s.Conn.RemoteAddr() }

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
  CHealthCheck
  CInputParameters
  CRunComputation
  CGetResult
)

type Parameter struct {
  Size uint32
  Data []byte
}

type InputParameters []*Parameter

func ReadParameters(sock Socket) (parameters InputParameters, n int, err error) {
  var pcount uint32

  // number of parameters
  err = binary.Read(sock, binary.BigEndian, &pcount)
  // TODO: handle error

  var nbytes, i uint32
  var nread int

  parameters = make(InputParameters, pcount, pcount)
  
  for i = 0; i < pcount; i++ {
    // parameter size : uint32
    err = binary.Read(sock, binary.BigEndian, &nbytes)
        
    data := make([]byte, nbytes, nbytes)
    nread, err = io.ReadFull(sock, data)

    // TODO: handle errors here
    if uint32(nread) == nbytes && err == nil {
      p := Parameter{nbytes, data}
      parameters[i] = &p
      n++
    } else {
      parameters[i] = nil
    }
  }

  return parameters, n, err
}
