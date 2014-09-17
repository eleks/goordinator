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
  input = flag.String("i", "input", "file with input parameters for computations")
  // 4 - fairly chosen by dice roll
  grind = flag.Int("n", 4, "grind number")
  logfile = flag.String("l", "worker.log", "absolute path to log file")
)


func main() {
  parseFlags()

  f, err := setupLogging()
  if err != nil {
    defer f.Close()
  }

  initConnection()

  canGetResults := make(chan bool, 1)
  startHealthcheck(canGetResults)

  commonParams, err := readCommonParameters("anyfile")
  checkFail(err)
  
  err = sendCommonParameters(commonParams)
  checkFail(err)
 
  err = computeTasks()
  checkFail(err)

  <- canGetResults
  err = sendCollectResultsRequest()
  if err != nil {
    return 
  }

  results := make(chan common.ComputationResult)
  resultsHandled := make(chan bool)
  go handleTaskResults(results)
  receiveResults(results)

  <- resultsHandled
  
  saveResults()
}

func checkFail(err error) {
  if err != nil {
    log.Fatal(err)
  }  
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

