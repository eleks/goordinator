package main

import (
  "../common"
  "net"
  "log"
  "encoding/binary"
  "container/heap"
)



type Coordinator struct {
  pool Pool
  client *Client
  commondata common.DataArray
  hash map[net.Addr]*Worker
  tasks chan common.Task
  tasks_count uint32
  worker_timeout chan HealthReporter
  client_timeout chan HealthReporter
  worker_quit chan bool
  client_quit chan bool
}

func (c *Coordinator)handleWorkerChannels(wch WorkerChannels) {
WorkerLoop:
  for {
    select {
    case w := <- wch.addworker: c.addWorker(w)
    case sock := <- wch.healthcheck_request: c.checkHealthWorker(sock)
    case task := <- c.tasks: c.dispatch(task)
    case sock := <- wch.gettask_request: c.sendNextTaskToWorker(sock)
    case <- c.worker_timeout: {
      // TODO: handle worker timeout (rebalance tasks)
      // TODO: distinguish worker and client
    }
    case <- c.worker_quit: {
      break WorkerLoop
      }
    }
  }
}

func (c *Coordinator)handleClientChannels(cch ClientChannels) {
ClientLoop:
  for {
    select {
    case cl := <- cch.addclient: c.addClient(cl)
    case sock := <- cch.healthcheck_request: c.checkHealthClient(sock)
    case sock := <- cch.readcommondata: c.readCommonData(sock)
    case sock := <- cch.runcomputation: c.runComputation(sock)
    case <- c.client_timeout: {
      // TODO: cleanup
    }
    case <- c.client_quit: {
      break ClientLoop
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

  go w.doWork()
}

func (c *Coordinator) addClient(cl *Client) {
  defer cl.CloseSock()
  
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

func (c *Coordinator) readCommonData(sock common.Socket) {
  defer sock.Close()

  parameters, n, err := common.ReadParameters(sock)
  if err != nil {
    log.Println(err)
  }

  c.commondata = parameters
}

func (c *Coordinator) runComputation(sock common.Socket) {
  // read all tasks parameters and create tasks
  var tcount, i uint32
  err := binary.Read(sock, binary.BigEndian, &tcount)
  // TODO: handle error

  task_id := 0

  for i = 0; i < tcount; i++ {
    parameters, n, err := common.ReadParameters(sock)

    if n == len(parameters) && err == nil {
      c.tasks <- common.Task{task_id, parameters}
      task_id++
    } else {
      log.Fatal(err)
    }
  }

  c.tasks_count = tcount
}

func (c *Coordinator) sendNextTaskToWorker(sock common.Socket) {
  addr := sock.RemoteAddr()

  w, present := c.hash[addr]
  if !present {
    log.Fatalf("sendNextTask: worker %v is not registered", addr)
  }
  
  go w.sendNextTask(sock)
}

func (c *Coordinator)dispatch(task common.Task) {
  w := heap.Pop(&c.pool).(*Worker)

  if w.pending < w.capacity {
    w.tasks <- task
    w.pending++    
  } else {
    // add same task again
    go func() {
      c.tasks <- task
    }()
  }

  heap.Push(&c.pool, w)
}

func (c *Coordinator)broadcast(task common.Task) {
  for _, w := range c.pool {
    w.tasks <- task
  }
}

func (c *Coordinator)completed(w *Worker) {
  
}
