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
  hash map[net.Addr]*Worker
  // buffered
  tasks chan common.Task
  workerTimeout chan HealthReporter
  clientTimeout chan HealthReporter
  workerQuit chan bool
  clientQuit chan bool
}

func (c *Coordinator)handleWorkerChannels(wch WorkerChannels) {
  log.Println("Coordinator: worker channels handling started")
  defer log.Println("Coordinator: worker channels handling finished")
  
WorkerLoop:
  for {
    select {
    case w := <- wch.addworker: c.addWorker(w)
    case sock := <- wch.healthcheckRequest: c.checkHealthWorker(sock)
    case newTask := <- c.tasks: c.dispatch(newTask)
    case sock := <- wch.gettaskRequest: c.sendNextTaskToWorker(sock)
    case hr := <- c.workerTimeout: c.workerTimeoutOccured(hr)
    case <- c.workerQuit: {
      break WorkerLoop
      }
    }
  }
}

func (c *Coordinator)handleClientChannels(cch ClientChannels) {
  log.Println("Coordinator: worker channels handling started")
  defer log.Println("Coordinator: worker channels handling finished")
  
ClientLoop:
  for {
    select {
    case cl := <- cch.addclient: c.addClient(cl)
    case sock := <- cch.healthcheckRequest: c.checkHealthClient(sock)
    case sock := <- cch.readcommondata: c.readCommonData(sock)
    case sock := <- cch.runcomputation: c.runComputation(sock)
    case <- c.clientTimeout: {
      // TODO: cleanup
    }
    case <- c.clientQuit: {
      break ClientLoop
      }
    }
  }
}

func (c *Coordinator)quit() {
  log.Println("Coordinator: posting quit signal to worker and client quit channels")
  c.workerQuit <- true
  c.clientQuit <- true
}

func (c *Coordinator) addWorker(w *Worker) {
  log.Printf("Coordinator: add worker with remote address %v \n", w.sock.RemoteAddr())
  
  heap.Push(&c.pool, *w)
  c.hash[w.sock.RemoteAddr()] = w

  go w.doWork()
}

func (c *Coordinator) addClient(cl *Client) {
  defer cl.CloseSock()

  log.Printf("Coordinator: add client with remote address %v \n", cl.sock.RemoteAddr())
  
  canAddClient := c.client == nil
  if canAddClient {
    c.client = cl
  }
  
  cl.replyInit(canAddClient)
}

func (c *Coordinator) checkHealthWorker(sock common.Socket){
  // TODO: add assert value,present = hash[addr]
  addr := sock.RemoteAddr()

  log.Printf("Worker healthcheck request from address: %v\n", addr)

  w, present := c.hash[addr]
  if !present {
    log.Fatalf("Healthcheck: worker with address %v is not registered", addr)
  }
  
  go checkHealth(w, c.workerTimeout)
}

func (c *Coordinator) checkHealthClient(sock common.Socket) {
  addr := sock.RemoteAddr()

  log.Printf("Client healthcheck request from address: %v\n", addr)

  cl, present := c.client, c.client != nil
  if !present {
    log.Fatalf("Healthcheck: client with address %v is not registered", addr)
  }

  go checkHealth(cl, c.clientTimeout)
}

func (c *Coordinator) readCommonData(sock common.Socket) {
  defer sock.Close()
  log.Println("Reading common data from client")
  defer log.Println("Common data has been read")

  parameters, _, err := common.ReadDataArray(sock)
  if err != nil {
    log.Println(err)
  }

  c.client.commondata = parameters
}

func (c *Coordinator) runComputation(sock common.Socket) {
  log.Println("Coordinator: reading specific parameters")

  // read all tasks parameters and create tasks
  var tcount, i uint32
  err := binary.Read(sock, binary.BigEndian, &tcount)
  // TODO: handle error
  if err != nil {
    log.Println(err)
  }

  log.Printf("Coordinator: going to receive %v tasks\n", tcount)

  var taskID int64

  for i = 0; i < tcount; i++ {
    parameters, n, err := common.ReadDataArray(sock)

    if n == len(parameters) && err == nil {
      c.tasks <- common.Task{taskID, parameters}
      taskID++
    } else {
      log.Fatal(err)
    }
  }

  c.client.tasksCount = tcount
}

func (c *Coordinator) sendNextTaskToWorker(sock common.Socket) {
  addr := sock.RemoteAddr()

  w, present := c.hash[addr]
  if !present {
    log.Fatalf("sendNextTask: worker %v is not registered", addr)
  }
  
  go w.sendNextTask(sock)
}

func (c *Coordinator) workerTimeoutOccured(hr HealthReporter) {
  addr := hr.GetSock().RemoteAddr()
  w, present := c.hash[addr]
  if !present {
    log.Fatal("Worker is not present in coordinator pool")
  }

  if w.pending > 0 {
    // rebalance all existing tasks
    for _, task := range w.activeTasks {
      go func() { c.tasks <- *task }()
    }
  }

  delete(c.hash, addr)
}

func (c *Coordinator)dispatch(task common.Task) {
  w := heap.Pop(&c.pool).(*Worker)

  // TODO: add assert pending >= 0
  
  if w.pending < w.capacity && w.pending >= 0 {
    go func(tasks chan common.Task, t) {tasks <- t} (w.tasks, task)
    w.pending++
  } else {
    // add same task again
    go func(tasks chan common.Task, t) {tasks <- t} (c.tasks, task)
  }

  heap.Push(&c.pool, w)
}

func (c *Coordinator)broadcast(task common.Task) {
  log.Println("Broadcasting task between workers")
  for _, w := range c.pool {
    w.tasks <- task
  }
}

func (c *Coordinator)completed(w *Worker) {
  
}
