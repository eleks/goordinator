package main

import (
  "../common"
  "net"
  "container/heap"
)

type healthCheckFunc func()

type Coordinator struct {
  pool Pool
  hash map[net.Addr]*Worker
  addworker chan Worker
  healthcheck chan net.Addr
  rmworker chan *Worker
  quit chan bool
}

func (c *Coordinator)handleChannels() {
Loop:
  for {
    select {
    case w := <- c.addworker:
      heap.Push(&c.pool, w)
    case <-c.quit:
      break Loop
    }
  }
}

func (c *Coordinator)dispatch(task common.Task) {
  
}

func (c *Coordinator)broadcast(task common.Task) {
  
}

func (c *Coordinator)completed(w *Worker) {
  
}
