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
  Parameters DataArray
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

type GenericData struct {
  Size uint32
  Data []byte
}

type DataArray []*GenericData

func ReadDataArray(sock Socket) (darray DataArray, n int, err error) {
  var pcount uint32

  // number of parameters
  err = binary.Read(sock, binary.BigEndian, &pcount)
  // TODO: handle error

  var nbytes, i uint32
  var nread int

  darray = make(DataArray, pcount, pcount)
  
  for i = 0; i < pcount; i++ {
    // parameter size : uint32
    err = binary.Read(sock, binary.BigEndian, &nbytes)
        
    data := make([]byte, nbytes, nbytes)
    nread, err = io.ReadFull(sock, data)

    // TODO: handle errors here
    if uint32(nread) == nbytes && err == nil {
      p := GenericData{nbytes, data}
      darray[i] = &p
      n++
    } else {
      darray[i] = nil
    }
  }

  return darray, n, err
}

func WriteDataArray(sock Socket, darray DataArray) error {
  err := binary.Write(sock, binary.BigEndian, uint32(len(darray)))
  // TODO: handle errors

  if err != nil {
    return err
  }

Loop:
  for _, p := range darray {
    err = binary.Write(sock, binary.BigEndian, p.Size)
    if err == nil {
      err = binary.Write(sock, binary.BigEndian, p.Data)
    }

    if err != nil {
      break Loop
    }
  }

  return err
}
