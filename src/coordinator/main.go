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
  worker_healthcheck := make(chan common.Socket)
  client_healthcheck := make(chan common.Socket)
  
  coordinator := &Coordinator{pool: make(Pool, 0, initialWorkerCount)}

  go coordinator.handleChannels(addworker, addclient, worker_healthcheck, client_healthcheck)
  
  handleClientsConnections(addclient, worker_healthcheck)
  handleWorkersConnections(addworker, client_healthcheck)
}

func parseFlags() {
  flag.Parse()

  if *wcap <= 0 || *wcap > maxWorkerTasksCapacity {
    log.Fatalf("worker tasks capacity must be between 0 and %d non-inclusively", maxWorkerTasksCapacity)
  }
}
