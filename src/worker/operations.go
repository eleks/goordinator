package main

import (
  "../common"
  "encoding/binary"
  "net"
  "log"
  "time"
)

func initConnection(cm ComputationManager) (err error) {
  reconnectTries := maxReconnectTries

reconnect:
  for ; reconnectTries > 0; reconnectTries-- {
    err = connectToCoordinator(cm)
    if err == nil {
      break reconnect
    }

    log.Printf("%v more tries left...\n", reconnectTries)
    time.Sleep(1 * time.Second)
  }

  return err
}

func connectToCoordinator(cm ComputationManager) (err error) {
  var conn net.Conn
  log.Printf("Connecting to coordinator %v...\n", *caddr)
    
  conn, err = net.Dial("tcp", *caddr)

  if err == nil {
    log.Println("Connected successfully to coordinator")
    binary.Write(conn, binary.BigEndian, common.WInit)
    var id uint32
    err = binary.Read(conn, binary.BigEndian, &id)
    if err == nil {
      cm.ID = id
    }
  } else {
    log.Println(err)
  }

  return err
}

func startHealthcheck(cm ComputationManager) {
  conn, err := net.Dial("tcp", *caddr)

  if err != nil {
    log.Printf("Unable to connect to coordinator.. exiting")
    return
  } else {
    log.Printf("Connected to coordinator. Sending healthcheck session request")
  }

  err = binary.Write(conn, binary.BigEndian, common.WHealthCheck)
  if err != nil {
    log.Printf("Unable to init healthcheck session")
    return
  }

  err = binary.Write(conn, binary.BigEndian, cm.ID)
  if err != nil {
    log.Printf("Unable to send worker ID")
    return
  }

  go healthcheckMainLoop(cm, conn)
}

func healthcheckMainLoop(cm ComputationManager, conn net.Conn) {
  infoChannel := make(chan uint32)

healthCheck:
  for {
    start := time.Now()
    
    cm.statusInfo <- infoChannel
    // uint32
    doneTasksCount := <- infoChannel
    binary.Write(conn, binary.BigEndian, doneTasksCount)

    var pending int
    // TODO: handle error
    err := binary.Read(conn, binary.BigEndian, &pending)

    if err == nil {
      go func(c ComputationManager, p int) {c.healthcheckResponse <- p}(cm, pending)
    } else {
      break healthCheck
    }

    common.SleepDifference(time.Since(start), 1.0)
  }
}
