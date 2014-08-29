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
  case common.WTaskCompeted:
  case common.WSendResult:
  }

  // wait until socket is processed
  <-sock.Done
  return nil
}

