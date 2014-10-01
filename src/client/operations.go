package main

import (
  "../common"
  "encoding/binary"
  "net"
  "log"
  "time"
)

func readCommonParameters() (params []common.Tasker, err error) {
  params = make([]common.Tasker, 1)

  objects := make([]common.SceneSerializer, 6)
  objects[0] = &common.SphereObject{
    Center: common.Vector3{0, -10004, -20},
    R: 10000,
    SurfaceColor: common.Vector3{0.2, 0.2, 0.2},
    EmissionColor: common.Vector3{},
    Transparency: 0,
    Reflection: 0,
  }

  objects[1] = &common.SphereObject{
    Center: common.Vector3{0, 0, -20},
    R: 4,
    SurfaceColor: common.Vector3{1.0, 0.32, 0.36},
    EmissionColor: common.Vector3{},
    Transparency: 0.5,
    Reflection: 1,
  }

  objects[2] = &common.SphereObject{
    Center: common.Vector3{5, -1, -15},
    R: 2,
    SurfaceColor: common.Vector3{0.9, 0.76, 0.46},
    EmissionColor: common.Vector3{},
    Transparency: 0,
    Reflection: 1,
  }

  objects[3] = &common.SphereObject{
    Center: common.Vector3{5, 0, -25},
    R: 3,
    SurfaceColor: common.Vector3{0.65, 0.77, 0.97},
    EmissionColor: common.Vector3{},
    Transparency: 0,
    Reflection: 1,
  }

  objects[4] = &common.SphereObject{
    Center: common.Vector3{-5.5, 0, -15},
    R: 3,
    SurfaceColor: common.Vector3{0.9, 0.9, 0.9},
    EmissionColor: common.Vector3{},
    Transparency: 0,
    Reflection: 1,
  }

  // light
  objects[5] = &common.SphereObject{
    Center: common.Vector3{0, 20, -30},
    R: 3,
    SurfaceColor: common.Vector3{},
    EmissionColor: common.Vector3{3, 3, 3},
    Transparency: 0,
    Reflection: 0,    
  }

  params[0] = &common.BeginSessionTasker{
    Width: 640,
    Height: 480,
    Angle: 60,
    SceneObjects: objects}

  return params, nil
}

func readRealParameters(filename string) (params []common.Tasker, err error) {
  // simulate reading for test
  h, w, d := 100, 100, 100

  data := make([]float32, h*w*d)

  for i:=0; i < h; i++ {
    for j:=0; j < w; j++ {
      for k:=0; k < d; k++ {
        if i == j && j == k {
          index := i*w+j + h*w*k
          data[index] = 1.0
        }
      }
    }
  }

  params = make([]common.Tasker, 1)
  params[0] = &common.TaskParameterFloat{
    Data: data,
    Dim1: uint32(h),
    Dim2: uint32(w),
    Dim3: uint32(d)}
  
  return params, nil
}

func initConnection() {
  log.Println("Waiting for a coordinator connection...")
  
Loop:
  for {
    log.Printf("Connecting to coordinator %v...\n", *caddr)
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

  log.Println("Reconnect loop finished...")
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
    second := time.After(1*time.Second)

    heartBeat := uint32(common.CIdle)
    binary.Write(conn, binary.BigEndian, heartBeat)

    var percentage int32
    // TODO: handle error
    err := binary.Read(conn, binary.BigEndian, &percentage)

    if err == nil {
      log.Printf("Percentage is %v", percentage)
      
      if percentage == 100 && !notified {
        log.Printf("Notifying abour results")
        canGetResults <- true
        notified = true
      }
    } else {
      log.Fatal(err)
    }

    <- second
  }
}

func sendCommonParameters(params []common.Tasker) error {
  conn, err := startCommonParamsConnection()
  if err != nil { return err }

  common.WriteTaskers(conn, params)
  
  if err == nil {
    log.Println("Common parameters have been sent successfully")
  } else  {
    log.Print("Failed to send part of generic data, error (%v)\n", err)
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

func computeTasks(parameters []common.Tasker, grindNumber int) error {
  conn, err := startComputeConnection()
  if err != nil { return err }

  bw := &common.BinWriter{W: conn}

  paramLength := uint32(len(parameters))
  log.Printf("Sending grinded (grind number %v) parameters to coordinator...", grindNumber)

  log.Printf("Sending batches count (%v)", grindNumber)
  bw.Write(uint32(grindNumber))

  for i:=0; i < grindNumber; i++ {
    bw.Write(paramLength)

    for j, p := range parameters {
      subtask := p.GetSubTask(uint32(i), uint32(grindNumber))
      log.Printf("Dumping subtask %v of parameter %v (size = %v)", j, i, subtask.GetBinarySize())
      
      bw.Write(subtask)

      if bw.Err != nil {
        log.Printf("Error while dumping subtask %v of parameter %v (%v)\n", i, j, bw.Err)
        return bw.Err
      }
    }
  }

  log.Println("Sending tasks loop finished")

  return bw.Err
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
    log.Printf("Unable to start main computation session (%v)\n", err)
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

func receiveResults(results chan *common.ComputationResult) {
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

func receiveResultsLoop(results chan *common.ComputationResult, finished chan bool) {
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

func getOneResult(results chan *common.ComputationResult, ping chan bool) error {
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
      go func(rs chan *common.ComputationResult, cr *common.ComputationResult) {rs <- cr} (results, &common.ComputationResult{gd, taskID})
    }
  }

  return nil
}

func handleTaskResults(results chan *common.ComputationResult, handled chan bool) {
saveLoop:
  for {
    select {
    case cr := <- results: saveTaskResult(cr)
    case <- time.After(1 * time.Minute): break saveLoop
    }
  }

  handled <- true
}

func saveTaskResult(cr *common.ComputationResult) {
  log.Printf("Received task result for task #%v with size %v", cr.ID, cr.Size)
  // TODO: implement
}

func saveResults() {
}
