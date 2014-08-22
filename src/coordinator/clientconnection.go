package main

import (
  "net"
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
  // tasks generation
  return nil
}
