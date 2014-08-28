package main

import (
  "../common"
  "net"
  "io"
  "log"
  "fmt"
)

func handleClientsConnections(coordinator *Coordinator) {
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

    go handleClient(conn)
  }
}

func handleClient(conn net.Conn) error {
  op_type := make([]byte, 1)

  _, err := io.ReadFull(conn, op_type)
  if err != nil {
    return err
  }

  optype := common.ClientOperation(op_type[0])
  switch optype {
  case common.CInitSession:
  case common.CInputParameters:
  case common.CRunComputation:
  case common.CGetResult:
  }
  // tasks generation
  return nil
}
