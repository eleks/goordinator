package main

import (
  "../common"
  "net"
  "log"
  "container/heap"
)

type healthCheckFunc func()

type Coordinator struct {
  pool Pool
  hash map[net.Addr]*Worker
  worker_timeout chan *Worker
  rmworker chan *Worker
  quit chan bool
}

func (c *Coordinator)handleChannels(addworker chan Worker, healthcheck_request chan common.Socket) {
Loop:
  for {
    select {
    case w := <- c.addworker: {
        heap.Push(&c.pool, w)
        c.hash[w.sock.RemoteAddr()] = &w
      }
    case sock := <- c.healthcheck_request: {
        // TODO: add assert value,present = hash[addr]
        addr := sock.RemoteAddr()
        w, present := c.hash[addr]
        if !present {
          log.Fatalf("Healthcheck: worker with address %v is not registered", addr)
        }

        go w.checkHealth(worker_timeout)
      }
    case w := <- worker_timeout: {
      // TODO: handle worker timeout (rebalance tasks)
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
