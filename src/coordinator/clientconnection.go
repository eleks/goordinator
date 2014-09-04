package main

import (
  "../common"
  "net"
  "io"
  "log"
  "fmt"
)

func handleClientsConnections(cch ClientChannels, client_quit chan bool) {
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

    // TODO: implement break
    go handleClient(sock, cch)
  }

  client_quit <- true
}

func handleClient(sock common.Socket, cch ClientChannels) error {
  log.Printf("Client connected from %v\n", sock.RemoteAddr())
  
  op_type := make([]byte, 1)

  _, err := io.ReadFull(sock, op_type)
  if err != nil {
    return err
  }

  optype := common.ClientOperation(op_type[0])
  switch optype {
  case common.CInitSession:
    cch.addclient <- &Client{sock: sock, status: common.CIdle}
  case common.CHealthCheck:
    cch.healthcheck_request <- sock
  case common.CInputParameters:
    cch.readcommondata <- sock
  case common.CRunComputation:
    cch.runcomputation <- sock
  case common.CGetResult:
  }

  log.Println("Waiting for client connection to finish")
  <-sock.Done
  log.Printf("Client disconnected from %v\n", sock.RemoteAddr())
  
  return nil
}
