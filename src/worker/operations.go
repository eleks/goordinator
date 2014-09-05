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

  info_channel := make(chan common.WorkerStatus)

  for {
    select {
    case <- ticker.C: go sendHealthCheck(cm, info_channel)
    }
  }
}

func sendHealthCheck(cm ComputationManager, info_channel chan common.WorkerStatus) {
  // TODO: handle error
  conn, _ := net.Dial("tcp", *caddr)

  cm.status_info <- info_channel
  health := <- info_channel
  binary.Write(conn, binary.BigEndian, health)

  var pending int
  // TODO: handle error
  err := binary.Read(conn, binary.BigEndian, pending)

  if err == nil {
    cm.healthcheck_response <- pending
  }
}
