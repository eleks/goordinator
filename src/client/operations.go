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

