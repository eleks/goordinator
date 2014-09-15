package main

import (
  "../common"
  "log"
  "io"
  "time"
  "encoding/binary"
)

type HealthReporter interface {
  GetStatus() interface{}
  GetStatusChannel() chan chan interface{}

  SetHealthStatus(status byte)
  GetHealthReply() interface{}

  GetID() uint32
}

func checkHealth(hr HealthReporter, sock common.Socket, timeout chan HealthReporter) {
  defer sock.CloseSock()

  healthcheck := make(chan uint32, 1)
  reply := make(chan interface{})
  done := make(chan bool, 1)

  go func() {
    var tasksDone uint32

  healthLoop:
    for {      
      err := binary.Read(sock, binary.BigEndian, &tasksDone)
      if err != nil {
        log.Fatal(err)
      }

      healthcheck <- tasksDone

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

  Loop:
  for {
    select {
    case status := <- healthcheck: {
      hr.SetHealthStatus(status)
      reply <- hr.GetHealthReply()
    }
    case hchannel := <- statusChannel: hchannel <- hr.GetStatus()
    case <- time.After(1 * time.Second): {
      // TODO: change timeout
      done <- true
      timeout <- hr
      break Loop
    }
    }
  }
}
