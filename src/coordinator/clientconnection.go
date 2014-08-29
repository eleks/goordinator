package main

import (
  "../common"
  "net"
  "io"
  "log"
  "fmt"
)

func handleClientsConnections(addclient chan<- *Client, healthcheck_request chan<- common.Socket) {
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
    
    go handleClient(sock, addclient, healthcheck_request)
  }
}

func handleClient(sock common.Socket, addclient chan<- *Client, healthcheck_request chan<- common.Socket) error {
  op_type := make([]byte, 1)

  _, err := io.ReadFull(sock, op_type)
  if err != nil {
    return err
  }

  optype := common.ClientOperation(op_type[0])
  switch optype {
  case common.CInitSession:
    addclient <- &Client{sock: sock, status: common.CIdle}
  case common.CHealthCheck:
    healthcheck_request <- sock
  case common.CInputParameters:
  case common.CRunComputation:
  case common.CGetResult:
  }

  <-sock.Done
  return nil
}
