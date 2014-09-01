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
  worker_timeout chan HealthReporter
  client_timeout chan HealthReporter
  worker_quit chan bool
  client_quit chan bool
}

func (c *Coordinator)handleWorkerChannels(wch WorkerChannels) {
Loop:
  for {
    select {
    case w := <- wch.addworker: c.addWorker(w)
    case sock := <- wch.healthcheck_request: c.checkHealthWorker(sock)
    case <- c.worker_timeout: {
      // TODO: handle worker timeout (rebalance tasks)
      // TODO: distinguish worker and client
    }
    case <- c.worker_quit: {
      break Loop
      }
    }
  }
}

func (c *Coordinator)handleClientChannels(cch ClientChannels) {
Loop:
  for {
    select {
    case cl := <- cch.addclient: c.addClient(cl)
    case sock := <- cch.healthcheck_request: c.checkHealthClient(sock)
    case <- c.client_timeout: {
      // TODO: cleanup
    }
    case <- c.client_quit: {
      break Loop
      }
    }
  }
}

func (c *Coordinator)quit() {
  c.worker_quit <- true
  c.client_quit <- true
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
  
  go checkHealth(w, c.worker_timeout)
}

func (c *Coordinator) checkHealthClient(sock common.Socket) {
  addr := sock.RemoteAddr()

  cl, present := c.client, c.client != nil
  if !present {
    log.Fatalf("Healthcheck: client with address %v is not registered", addr)
  }

  go checkHealth(cl, c.client_timeout)
}

func (c *Coordinator)dispatch(task common.Task) {
  
}

func (c *Coordinator)broadcast(task common.Task) {
  
}

func (c *Coordinator)completed(w *Worker) {
  
}
