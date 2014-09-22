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

  log.Println("Reconnect loop finished...")

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
      log.Printf("Received id #%v from coordinator", id)
    } else {
      log.Printf("Error on receiving Id (%v)", err)
    }
  } else {
    log.Printf("Error on dialing coordinator (%v)\n", err)
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
    log.Printf("Unable to init healthcheck session (%v)\n", err)
    return
  }

  err = binary.Write(conn, binary.BigEndian, cm.ID)
  if err != nil {
    log.Printf("Unable to send worker ID (%v)", err)
    return
  }

  go healthcheckMainLoop(cm, conn)
}

func healthcheckMainLoop(cm ComputationManager, conn net.Conn) {
  infoChannel := make(chan int32)

  log.Println("Healthcheck main loop started")
  
healthCheck:
  for {
    second := time.After(1*time.Second)
    
    cm.statusInfo <- infoChannel
    // int32
    doneTasksCount := <- infoChannel
    binary.Write(conn, binary.BigEndian, doneTasksCount)

    var pending int32
    err := binary.Read(conn, binary.BigEndian, &pending)

    if err == nil {
      log.Printf("Pending %v tasks", pending)
      go func(c ComputationManager, p int32) {c.healthcheckResponse <- p}(cm, pending)
    } else {
      log.Printf("Error while healthcheck, %v", err)
      break healthCheck
    }

    <- second
  }

  log.Printf("Healthcheck loop finished")
}
