package main

import (
  "../common"
  "fmt"
  "log"
  "net"
  "encoding/binary"
)

func handleWorkersConnections(wch WorkerChannels) {
  listener, err := net.Listen("tcp", *lwaddr)

  if err != nil {
    log.Fatal(err)
  }

  // TODO: implement break
  for {
    conn, err := listener.Accept()
    if err != nil {
      fmt.Println(err)
      continue
    }

    sock := common.Socket{conn, make(chan bool)}
    
    go handleWorker(sock, wch)
  }
}

func handleWorker(sock common.Socket, wch WorkerChannels) error {
  log.Printf("Worker connected from %v\n", sock.RemoteAddr())
  
  var opType byte
  var workerID uint32

  err := binary.Read(sock, binary.BigEndian, &opType)
  if err != nil {
    return err
  }

  optype := common.WorkerOperation(opType)
  switch optype {
  case common.WInit:
    wch.addworker <- &Worker{
      tasks: make(chan common.Task),
      stop: make(chan bool),
      cinfo: make(chan interface{}),
      ccinfo: make(chan chan interface{}),
      ID: 0}
    nextID := <- wch.nextID
    err = binary.Write(sock, binary.BigEndian, nextID)
    go sock.Close()
  case common.WHealthCheck:
    err = binary.Read(sock, binary.BigEndian, &workerID)
    if err == nil {
      wch.healthcheckRequest <- WCInfo{workerID, sock}
    }
  case common.WGetTask:
    err = binary.Read(sock, binary.BigEndian, &workerID)
    if err == nil {
      wch.gettaskRequest <- WCInfo{workerID, sock}
    }
  case common.WTaskCompeted: // can be implemented on demand
  case common.WSendResult:
    wch.taskresult <- sock
  }

  // wait until socket is processed
  
  <-sock.Done
  log.Printf("Worker disconnected from %v\n", sock.RemoteAddr())
  return nil
}
