package main

import (
  "../common"
  "encoding/binary"
  "net"
  "log"
  "time"
)

func initWorker(cm ComputationManager) (err error) {
  var conn net.Conn
  reconnectTries := maxReconnectTries
  
  for ; reconnectTries > 0; reconnectTries-- {
    log.Printf("Connecting to coordinator %v...\n", *caddr)
    
    conn, err = net.Dial("tcp", *caddr)

    if err == nil {
      log.Println("Connected successfully")
      binary.Write(conn, binary.BigEndian, common.WInit)
    } else {
      log.Println(err)
    }

    log.Printf("%v more tries left...\n", reconnectTries)
    time.Sleep(1 * time.Second)
  }

  return err
}

func startHealthcheck(cm ComputationManager) {
  ticker := time.NewTicker(1 * time.Second)

  infoChannel := make(chan common.WorkerStatus)

  for {
    select {
    case <- ticker.C: go sendHealthCheck(cm, infoChannel)
    }
  }
}

func sendHealthCheck(cm ComputationManager, infoChannel chan common.WorkerStatus) {
  // TODO: handle error
  conn, _ := net.Dial("tcp", *caddr)

  cm.statusInfo <- infoChannel
  health := <- infoChannel
  binary.Write(conn, binary.BigEndian, health)

  var pending int
  // TODO: handle error
  err := binary.Read(conn, binary.BigEndian, pending)

  if err == nil {
    cm.healthcheckResponse <- pending
  }
}
