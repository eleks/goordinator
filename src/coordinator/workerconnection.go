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

    go handleWorker(conn, addworker)
  }
}

func handleWorker(conn net.Conn, addworker chan<- Worker) error {
  defer conn.Close()

  op_type := make([]byte, 1)

  _, err := io.ReadFull(conn, op_type)
  if err != nil {
    return err
  }

  optype := common.WorkerOperation(op_type[0])
  switch optype {
  case common.WInit:
    addworker <- Worker{addr: conn.RemoteAddr(), tasks: make(chan common.Task)}
    // TODO: send response
  case common.WHealthCheck:
    
  case common.WGetTask:
  case common.WTaskCompeted:
  case common.WSendResult:
  }
}

