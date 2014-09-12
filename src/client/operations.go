package main

import (
  "../common"
  "encoding/binary"
  "net"
  "log"
  "time"
)

func initConnection() {
  var connected bool

  log.Println("Waiting for a coordinator connection...")
  
Loop:
  for {
    conn, err := net.Dial("tcp", *caddr)

    if err == nil {
      log.Println("Connected successfully to coordinator")
      binary.Write(conn, binary.BigEndian, common.CInitSession)

      var success byte
      err = binary.Read(conn, binary.BigEndian, &success)

      connected = err == nil && success == 1
      break Loop
    }

    time.Sleep(2 * time.Second)
  }
}

func startHealthcheck() {
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

  go healthcheckMainLoop(cm, conn)
}

func healthcheckMainLoop(cm ComputationManager, conn net.Conn) {
    infoChannel := make(chan common.ClientStatus)
  
healthCheck:
  for {
    start := time.Now()
    
    cm.statusInfo <- infoChannel
    health := <- infoChannel
    binary.Write(conn, binary.BigEndian, health)

    var percentage uint32
    // TODO: handle error
    err := binary.Read(conn, binary.BigEndian, &percentage)

    if err == nil {
      // TODO: do smth with health reply
    } else {
      break healthCheck
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

  for i, t := range params {
    buf, err = t.GobEncode()
    if err == nil {
      gd := common.GenericData{uint32(len(buf)), buf}
      err = common.WriteGenericData(conn, gd)
    }

    if err != nil {
      log.Println("Failed to send part of generic data")
      log.Prinln(err)
    }
  }

  if err == nil {
    log.Prinln("Common parameters have been sent successfully")
  }

  return err
}

func startCommonParamsConnection() (conn net.Conn, err error) {
  conn, err = net.Dial("tcp", *caddr)

  if err != nil {
    log.Printf("Unable to connect to coordinator.. exiting")
    return err
  } else {
    log.Printf("Sending common parameters")
  }

  err = binary.Write(conn, binary.BigEndian, common.CInputParameters)
  if err != nil {
    log.Printf("Unable to init healthcheck session")
    return err
  }

  return nil
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

func startComputeConnection() (net.Conn, error) {
  conn, err = net.Dial("tcp", *caddr)
  
  if err != nil {
    log.Printf("Unable to connect to coordinator.. exiting")
    return err
  } else {
    log.Printf("Sending main parameters")
  }

  err = binary.Write(conn, binary.BigEndian, common.CRunComputation)
  if err != nil {
    log.Printf("Unable to start main computation session")
    return err
  }

  return nil
}

