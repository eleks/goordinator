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
    case w := <- wch.addworker: c.addWorker(w)
    case cl := <- cch.addclient: c.addClient(cl)
    case sock := <- wch.healthcheck_request: c.checkHealthWorker(sock)
    case sock := <- cch.healthcheck_request: c.checkHealthClient(sock)
    case w := <- c.worker_timeout: {
      // TODO: handle worker timeout (rebalance tasks)
      // TODO: distinguish worker and client
    }
    case <-c.quit: {
      break Loop
      }
    }
  }
}

func (c *Coordinator) addWorker(w *Worker) {
  heap.Push(&c.pool, *w)
  c.hash[w.sock.RemoteAddr()] = w
}

func (c *Coordinator) addClient(cl *Client) {
  canAddClient := c.client == nil
  if canAddClient {
    c.client = cl
  }
  
  cl.replyInit(canAddClient)
}

func (c *Coordinator) checkHealthWorker(sock common.Socket){
  // TODO: add assert value,present = hash[addr]
  addr := sock.RemoteAddr()

  w, present := c.hash[addr]
  if !present {
    log.Fatalf("Healthcheck: worker with address %v is not registered", addr)
  }
  
  if hr, ok := (*w).(HealthReporter); ok {  
    go checkHealth(hr, c.worker_timeout)
  }
}

func (c *Coordinator) checkHealthClient(sock common.Socket) {
  addr := sock.RemoteAddr()

  cl, present := c.client, c.client != nil
  if !present {
    log.Fatalf("Healthcheck: client with address %v is not registered", addr)
  }

  if hr, ok := (*cl).(HealthReporter); ok {  
    go checkHealth(hr, c.client_timeout)
  }
}

func (c *Coordinator)dispatch(task common.Task) {
  
}

func (c *Coordinator)broadcast(task common.Task) {
  
}

func (c *Coordinator)completed(w *Worker) {
  
}
