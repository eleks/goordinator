package main

import (
  "flag"
  "log"
  "os"
  "fmt"
  "time"
)

// Flags
var (
  caddr = flag.String("c", "0.0.0.0:4321", "the address of coordinator")
  logfile = flag.String("l", "worker.log", "absolute path to log file")
)


func main() {
  parseFlags()

  f, err := setupLogging()
  if err != nil {
    defer f.Close()
  }

  initConnection()

  startHealthcheck()

  sendCommonParameters()

  computeTasks()

  getResults()

  saveResults()
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
  log.Println("Client log started")
  
  return f, err
}

