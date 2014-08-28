package main

import (
  "../common"
  "flag"
  "log"
)

// Flags
var (
  lwaddr = flag.String("w", "0.0.0.0:4321", "the address to listen on for workers connections")
  lcaddr = flag.String("c", "0.0.0.0:4322", "the address to listen on for clients connections")
  wcap = flag.Int("p", 5, "worker tasks capacity in range (0, 100)")
)

const (
  maxWorkerTasksCapacity = 100
  initialWorkerCount = 20
)

func main() {
  parseFlags()

  addworker := make(chan *Worker)
  addclient := make(chan *Client)
  healthcheck_request := make(chan common.Socket)
  
  coordinator := &Coordinator{pool: make(Pool, 0, initialWorkerCount)}

  go coordinator.handleChannels(addworker, addclient, healthcheck_request)
  
  handleClientsConnections(addclient)
  handleWorkersConnections(addworker, healthcheck_request)
}

func parseFlags() {
  flag.Parse()

  if *wcap <= 0 || *wcap > maxWorkerTasksCapacity {
    log.Fatalf("worker tasks capacity must be between 0 and %d non-inclusively", maxWorkerTasksCapacity)
  }
}
