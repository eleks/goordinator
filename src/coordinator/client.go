package main

import (
  "../common"
  "log"
  "io"
  "time"
)

type Client struct {
  sock common.Socket
  common_data []common.Parameter
  status common.ClientStatus
}

func (c *Client) replyInit(success bool) {
  defer c.sock.Close()

  connection_status_buf := make([]byte, 1)
  if success {
    connection_status_buf[0] = 1
  }
  else {
    connection_status_buf[0] = 0
  }

  c.sock.Write(connection_status_buf)
}

func (c *Client) checkHealth(timeout chan *Client) {
  defer c.sock.Close()

  healthcheck := make(chan bool, 1)
  reply := make(chan common.ClientStatus)
  done := make(chan bool, 1)

  go func() {
    client_status_buf := make([]byte, 1)

  healthLoop:
    for {      
      _, err := io.ReadFull(c.sock, client_status_buf)
      if err != nil {
        // TODO: send errors to channel
        log.Fatal(err)
        break healthLoop
      }

      healthcheck <- true

      select {
      case client_status := <- reply: {
        client_status_buf[0] = byte(client_status)
        _, err := c.sock.Write(client_status_buf)
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

  Loop:
  for {
    select {
    case status := <- healthcheck: {
      reply <- c.status
    }
    case <- time.After(1 * time.Second): {      
      // TODO: change timeout
      done <- true
      timeout <- c
      break Loop
    }
    }
  }

}
