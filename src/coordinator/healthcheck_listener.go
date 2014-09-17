package main

import (
  "../common"
  "log"
  "time"
  "encoding/binary"
)

type HealthReporter interface {
  GetStatus() interface{}
  GetStatusChannel() chan chan interface{}

  SetHealthStatus(status uint32)
  GetHealthReply() interface{}

  GetID() uint32

  GetResultsFlagChannel() chan bool
  SetGetResultsFlag()
}

func checkHealth(hr HealthReporter, sock common.Socket, timeout chan HealthReporter) {
  defer sock.Close()

  healthcheck := make(chan uint32, 1)
  reply := make(chan interface{})
  done := make(chan bool, 1)

  go func() {
    // means tasksDone for worker
    var heartBeat uint32
    
  healthLoop:
    for {      
      err := binary.Read(sock, binary.BigEndian, &heartBeat)
      if err != nil {
        log.Fatal(err)
      }

      healthcheck <- heartBeat

      select {
      case healthReply := <- reply: {
        err := binary.Write(sock, binary.BigEndian, healthReply)
        if err != nil {
          // TODO: send errors to channel
          log.Fatal(err)
          break healthLoop
        }
      }
      case <- done: break healthLoop
      }
    }
  }()

  statusChannel := hr.GetStatusChannel()
  resultsFlagChannel := hr.GetResultsFlagChannel()

  Loop:
  for {
    select {
    case status := <- healthcheck: {
      hr.SetHealthStatus(status)
      reply <- hr.GetHealthReply()
    }
    case hchannel := <- statusChannel: hchannel <- hr.GetStatus()
    case <- resultsFlagChannel: hr.SetGetResultsFlag()
    case <- time.After(1 * time.Second): {
      // TODO: change timeout
      done <- true
      timeout <- hr
      break Loop
    }
    }
  }
}
