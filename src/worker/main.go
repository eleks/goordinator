package main

import (
  "log"
  "fmt"
  "os"
  "flag"
)

// Flags
var (
  caddr = flag.String("c", "0.0.0.0:4321", "the address of coordinator")
  logfile = flag.String("l", "worker.log", "absolute path to log file")
)

const (
  maxReconnectTries = 100
  defaultBufferLength = 10
)

func main() {
  parseFlags()

  f, err := setupLogging()
  if err != nil {
    defer f.Close()
  }

  cm := ComputationManager{
    healthcheckResponse: make(chan int),
    statusInfo: make(chan chan common.WorkerStatus),
    pendingTasksCount: 0,
    tasks: make(chan common.Task),
    results: make(map[int]common.ComputationResult)
    chResults: make(chan common.ComputationResult)
    status: common.WReady
    stopMessages: make(chan chan error)
    sendingMode: false
    // buffered
    stopComputations: make(chan bool, defaultBufferLength)
  }

  go cm.processTasks()

  err = initConnection(cm)
  go startHealthcheck(cm)

  cm.handleCommands()
}

func parseFlags() {
  flag.Parse()
}

func setupLogging() (f *os.File, err error) {
  f, err = os.OpenFile(*logfile, os.O_RDWR | os.O_CREATE | os.O_APPEND, 0666)
  if err != nil {
    fmt.Println("error opening file: %v", *logfile)
    return nil, err
  }

  log.SetOutput(f)
  log.Println("------------------------------")
  log.Println("Worker log started")
  
  return f, err
}

