package main

import (
  "../common"
  "net"
  "log"
  "container/heap"
)

type Coordinator struct {
  pool Pool
  client *Client
  hash map[net.Addr]*Worker
  timeout chan *HealthReporter
  rmworker chan *Worker
  quit chan bool
}

func (c *Coordinator)handleChannels(addworker chan *Worker, addclient chan *Client, healthcheck_request chan HealthReporter) {
Loop:
  for {
    select {
    case w := <- addworker: {
        heap.Push(&c.pool, *w)
        c.hash[w.sock.RemoteAddr()] = w
      }
    case cl := <- addclient: {
      canAddClient := c.client == nil
      if canAddClient {
        c.client = cl
      }
      
      cl.replyInit(canAddClient)
    }
    case hr := <- healthcheck_request: {
        // TODO: add assert value,present = hash[addr]
      sock := hr.GetSock()
      addr := sock.RemoteAddr()
      // TODO: distinguish worker and client hash
      w, present := c.hash[addr]
      if !present {
        log.Fatalf("Healthcheck: worker with address %v is not registered", addr)
      }

      go checkHealth(hr, c.timeout)
    }
    case w := <- c.timeout: {
      // TODO: handle worker timeout (rebalance tasks)
      // TODO: distinguish worker and client
    }
    case <-c.quit: {
      break Loop
      }
    }
  }
}

func (c *Coordinator)dispatch(task common.Task) {
  
}

func (c *Coordinator)broadcast(task common.Task) {
  
}

func (c *Coordinator)completed(w *Worker) {
  
}
