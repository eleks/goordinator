package main

import (
  "../common"
  "io"
)

type HealthReporter interface {
  GetStatusRequestChannel() chan bool
  GetStatus() interface {}
  GetStatusChannel() chan interface{}
  SetHealthStatus(status byte)
  GetHealthReply() interface{}

  CloseSock()
  GetSock() common.Socket
}

func checkHealth(hr HealthReporter, timeout chan *HealthReporter) {
  defer hr.CloseSock()

  healthcheck := make(chan byte, 1)
  reply := make(chan interface{})
  done := make(chan bool, 1)

  go func() {
    input_buf := make([]byte, 1)
    sock := hr.GetSock()

  healthLoop:
    for {      
      _, err := io.ReadFull(sock, input_buf)
      if err != nil {
        // TODO: send errors to channel
        log.Fatal(err)
        break healthLoop
      }

      health_status := input_buf[0]
      healthcheck <- health_status

      select {
      case health_reply := <- reply: {
        err := binary.Write(sock, binary.BigEndian, health_reply)
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

  request_status_channel := hr.GetStatusRequestChannel()
  status_channel := hr.GetStatusChannel()

  Loop:
  for {
    select {
    case status := <- healthcheck: {
      hr.SetHealthStatus(status)
      reply <- hr.GetHealthReply()
    }
    case <- request_status_channel: status_channel <- hr.GetStatus()
    case <- time.After(1 * time.Second): {
      // TODO: change timeout
      done <- true
      timeout <- hr
      break Loop
    }
    }
  }
}
