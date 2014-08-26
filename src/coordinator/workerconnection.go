package main

import (
  "../common"
  "fmt"
  "log"
  "net"
  "io"
)

func handleWorkersConnections(addworker chan<- Worker) {
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

    sock = common.Socket {conn, make(chan bool)}

    go handleWorker(sock, addworker)
  }
}

func handleWorker(sock common.Socket, addworker chan<- Worker, healthcheck chan<- common.Socket) error {
  op_type := make([]byte, 1)

  _, err := io.ReadFull(conn, op_type)
  if err != nil {
    return err
  }

  optype := common.WorkerOperation(op_type[0])
  switch optype {
  case common.WInit:
    addworker <- Worker{sock: sock, tasks: make(chan common.Task)}
    // TODO: send response
  case common.WHealthCheck:
    healthcheck_request <- sock    
  case common.WGetTask:
  case common.WTaskCompeted:
  case common.WSendResult:
  }

  // wait until socket is processed
  <-sock.done
}

