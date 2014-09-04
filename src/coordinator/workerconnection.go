package main

import (
  "../common"
  "fmt"
  "log"
  "net"
  "io"
)

func handleWorkersConnections(wch WorkerChannels, worker_quit chan bool) {
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

  worker_quit <- true
}

func handleWorker(sock common.Socket, wch WorkerChannels) error {
  log.Printf("Worker connected from %v\n", sock.RemoteAddr())
  
  op_type := make([]byte, 1)

  _, err := io.ReadFull(sock, op_type)
  if err != nil {
    return err
  }

  optype := common.WorkerOperation(op_type[0])
  switch optype {
  case common.WInit:
    wch.addworker <- &Worker{sock: sock, tasks: make(chan common.Task)}
    // TODO: send response
  case common.WHealthCheck:
    wch.healthcheck_request <- sock
  case common.WGetTask:
    wch.gettask_request <- sock
  case common.WTaskCompeted:
  case common.WSendResult:
  }

  // wait until socket is processed
  
  <-sock.Done
  log.Printf("Worker disconnected from %v\n", sock.RemoteAddr())
  return nil
}

