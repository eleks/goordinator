package main

import (
  "../common"
  "flag"
  "net"
  "log"
  "fmt"
  "io"
)

// Flags
var (
  laddr = flag.String("l", "0.0.0.0:4321", "the address to listen on")
  wcap = flag.Int("w", 5, "worker tasks capacity in range (0, 100)")
)

const (
  maxWorkerTasksCapacity = 100
)

func main() {
  parseFlags()
  handleConnections()
}

func parseFlags() {
  flag.Parse()

  if *wcap <= 0 || *wcap > maxWorkerTasksCapacity {
    log.Fatalf("worker tasks capacity must be between 0 and %d non-inclusively", maxWorkerTasksCapacity)
  }
}

func handleConnections() {
  listener, err := net.Listen("tcp", *laddr)

  if err != nil {
    log.Fatal(err)
  }

  for {
    conn, err := listener.Accept()
    if err != nil {
      fmt.Println(err)
      continue
    }

    go handleConnection(conn)
  }
}

func handleConnection(conn net.Conn) error {
  conn_type := make([]byte, 1)

  for {
    _, err := io.ReadFull(conn, conn_type)
    if err != nil {
      return err
    }

    ctype := common.ConnectionType(conn_type[0])
    switch ctype {
    case common.ClientConnection:
      handleClient(conn)
    case common.WorkerConnection:
      handleWorker(conn)
    }
  }
}

func handleClient(conn net.Conn) {
  // tasks generation
}

func handleWorker(conn net.Conn) {
  
}
