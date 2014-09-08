package main

import (
  "../common"
  "fmt"
  "log"
  "net"
  "io"
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
  
  opType := make([]byte, 1)

  _, err := io.ReadFull(sock, opType)
  if err != nil {
    return err
  }

  optype := common.WorkerOperation(opType[0])
  switch optype {
  case common.WInit:
    wch.addworker <- &Worker{sock: sock, tasks: make(chan common.Task)}
    // TODO: send response
  case common.WHealthCheck:
    wch.healthcheckRequest <- sock
  case common.WGetTask:
    wch.gettaskRequest <- sock
  case common.WTaskCompeted:
  case common.WSendResult:
  }

  // wait until socket is processed
  
  <-sock.Done
  log.Printf("Worker disconnected from %v\n", sock.RemoteAddr())
  return nil
}

