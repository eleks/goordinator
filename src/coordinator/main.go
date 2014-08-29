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

  worker_channels := WorkerChannels{make(chan *Worker), make(chan common.Socket), make(chan *Worker)}
  client_channels := ClientChannels{make(chan *Client), make(chan common.Socket), make(chan *Client)}
  
  coordinator := &Coordinator{pool: make(Pool, 0, initialWorkerCount)}

  go coordinator.handleChannels(worker_channels, client_channels)
  
  handleClientsConnections(client_channels)
  handleWorkersConnections(worker_channels)
}

func parseFlags() {
  flag.Parse()

  if *wcap <= 0 || *wcap > maxWorkerTasksCapacity {
    log.Fatalf("worker tasks capacity must be between 0 and %d non-inclusively", maxWorkerTasksCapacity)
  }
}
