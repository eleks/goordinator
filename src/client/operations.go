package main

import (
  "../common"
  "encoding/binary"
  "net"
  "log"
  "time"
)

func readCommonParameters(filename string) (params []Tasker, err error) {
  return nil, nil
}

func readRealParameters(filename string) (params []Tasker, err error) {
  return nil, nil
}

func initConnection() {
  log.Println("Waiting for a coordinator connection...")
  
Loop:
  for {
    conn, err := net.Dial("tcp", *caddr)

    if err == nil {
      log.Println("Connected successfully to coordinator")
      binary.Write(conn, binary.BigEndian, common.CInitSession)

      var success byte
      err = binary.Read(conn, binary.BigEndian, &success)

      if err == nil && success == 1 {
        break Loop
      }
    }

    time.Sleep(2 * time.Second)
  }
}

func startHealthcheck(canGetResults chan bool) {
  conn, err := net.Dial("tcp", *caddr)

  if err != nil {
    log.Printf("Unable to connect to coordinator.. exiting")
    return
  } else {
    log.Printf("Connected to coordinator. Sending healthcheck session request")
  }

  err = binary.Write(conn, binary.BigEndian, common.CHealthCheck)
  if err != nil {
    log.Printf("Unable to init healthcheck session")
    return
  }

  go healthcheckMainLoop(conn, canGetResults)
}

func healthcheckMainLoop(conn net.Conn, canGetResults chan bool) {
  notified := false
  
  for {
    start := time.Now()
    
    binary.Write(conn, binary.BigEndian, common.CIdle)

    var percentage uint32
    // TODO: handle error
    err := binary.Read(conn, binary.BigEndian, &percentage)

    if err == nil {
      if percentage == 100 && !notified {
        canGetResults <- true
        notified = true
      }
    } else {
      log.Fatal(err)
    }

    common.SleepDifference(time.Since(start), 1.0)
  }
}

func sendCommonParameters(params []Tasker) error {
  conn, err := startCommonParamsConnection()
  if err != nil { return err }
  
  p := uint32(len(params))
  err = binary.Write(conn, binary.BigEndian, p)
  if err != nil { return err }

  var buf []byte

  for _, t := range params {
    buf, err = t.GobEncode()
    if err == nil {
      gd := common.GenericData{uint32(len(buf)), buf}
      err = gd.Write(conn)
    }

    if err != nil {
      log.Println("Failed to send part of generic data")
      log.Println(err)
    }
  }

  if err == nil {
    log.Println("Common parameters have been sent successfully")
  }

  return err
}

func startCommonParamsConnection() (conn net.Conn, err error) {
  conn, err = net.Dial("tcp", *caddr)

  if err != nil {
    log.Printf("Unable to connect to coordinator.. exiting")
    return nil, err
  } else {
    log.Printf("Sending common parameters")
  }

  err = binary.Write(conn, binary.BigEndian, common.CInputParameters)
  if err != nil {
    log.Printf("Unable to init healthcheck session")
    return nil, err
  }

  return conn, nil
}

func computeTasks(parameters []Tasker, grindNumber int) error {
  conn, err := startComputeConnection()
  if err != nil { return err }

  log.Println("Sending grinded parameters to coordinator")

  for i:=0; i < grindNumber; i++ {
    binary.Write(conn, binary.BigEndian, uint32(len(parameters)))

    for _, p := range parameters {
      subtask := p.GetSubTask(i, grindNumber)
      err = subtask.Dump(conn)

      if err != nil {
        return err
      }
    }
  }

  return nil
}

func startComputeConnection() (conn net.Conn, err error) {
  conn, err = net.Dial("tcp", *caddr)
  
  if err != nil {
    log.Printf("Unable to connect to coordinator.. exiting")
    return nil, err
  } else {
    log.Printf("Sending main parameters")
  }

  err = binary.Write(conn, binary.BigEndian, common.CRunComputation)
  if err != nil {
    log.Printf("Unable to start main computation session")
    return nil, err
  }

  return conn, nil
}

func sendCollectResultsRequest() error {
  conn, err := net.Dial("tcp", *caddr)
  
  if err == nil {
    err = binary.Write(conn, binary.BigEndian, common.CCollectResults)
  }

  if err != nil {
    log.Printf("Failed to ask coordinator to collect results")
  }

  return err
}

func receiveResults(results chan common.ComputationResult) {
  // in this version
  queues_number := 2
  
  waitAll := make([]chan bool, queues_number)
  for i := range waitAll { waitAll[i] = make(chan bool) }

  for i := range waitAll {
    go receiveResultsLoop(results, waitAll[i])
  }

  // actually, wait all of them
  for i := range waitAll { <- waitAll[i] }
}

func receiveResultsLoop(results chan common.ComputationResult, finished chan bool) {
  ping := make(chan bool)
getResults:
  for {
    go getOneResult(results, ping)

    // wait for coordinator connection
    select {
    case <- ping:
    case <- time.After(5*time.Second):
      break getResults
    }

    // wait for reading task itself
    select {
    case <- ping:
      // TODO: make these numbers constants
    case <- time.After(1 * time.Minute):
      break getResults
    }
  }

  finished <- true
}

func getOneResult(results chan common.ComputationResult, ping chan bool) error {
  conn, err := net.Dial("tcp", *caddr)
  if err != nil {
    return err
  }

  err = binary.Write(conn, binary.BigEndian, common.CGetResult)
  if err != nil {
    return err
  }

  ping <- true

  var taskID int64
  err = binary.Read(conn, binary.BigEndian, &taskID)

  var gd common.GenericData
  
  if err == nil {
    gd, err = common.ReadGenericData(conn)
    ping <- true

    if err == nil {
      go func(rs chan common.ComputationResult, cr common.ComputationResult) {rs <- cr} (results, common.ComputationResult{gd, taskID})
    }
  }

  return nil
}

func handleTaskResults(results chan common.ComputationResult, handled chan bool) {
saveLoop:
  for {
    select {
    case cr := <- results: saveTaskResult(cr)
    case <- time.After(1 * time.Minute): break saveLoop
    }
  }

  handled <- true
}

func saveTaskResult(cr common.ComputationResult) {
  // TODO: implement
}

func saveResults() {
}
