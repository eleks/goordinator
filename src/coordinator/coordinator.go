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
  worker_timeout chan *HealthReporter
  client_timeout chan *HealthReporter
  rmworker chan *Worker
  quit chan bool
}

func (c *Coordinator)handleChannels(wch WorkerChannels, cch ClientChannels) {
Loop:
  for {
    select {
    case w := <- wch.addworker: {
        heap.Push(&c.pool, *w)
        c.hash[w.sock.RemoteAddr()] = w
      }
    case cl := <- cch.addclient: {
      canAddClient := c.client == nil
      if canAddClient {
        c.client = cl
      }
      
      cl.replyInit(canAddClient)
    }
    case sock := <- wch.healthcheck_request: {
      // TODO: add assert value,present = hash[addr]
      addr := sock.RemoteAddr()

      w, present := c.hash[addr]
      if !present {
        log.Fatalf("Healthcheck: worker with address %v is not registered", addr)
      }

      go checkHealth(w, c.worker_timeout)
    }
    case sock := <- cch.healthcheck_request: {
      addr := sock.RemoteAddr()

      cl, present := c.client, c.client != nil
      if !present {
        log.Fatalf("Healthcheck: client with address %v is not registered", addr)
      }

      go checkHealth(cl, c.client_timeout)
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
