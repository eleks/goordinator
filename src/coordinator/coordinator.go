package main

import (
  "../common"
  "log"
  "encoding/binary"
  "container/heap"
)

type Coordinator struct {
  pool Pool
  client *Client
  hash map[uint32]*Worker
  lastWorkerID uint32
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
    case w := <- wch.addworker: c.addWorker(w, wch.nextID)
    case wci := <- wch.healthcheckRequest: c.checkHealthWorker(wci)
    case newTask := <- c.tasks: c.dispatch(newTask)
    case wci := <- wch.gettaskRequest: c.sendNextTaskToWorker(wci)
    case hr := <- c.workerTimeout: c.workerTimeoutOccured(hr)
    case <- c.workerQuit: {
      break WorkerLoop
      }
    }
  }
}

func (c *Coordinator)handleClientChannels(cch ClientChannels) {
  log.Println("Coordinator: client channels handling started")
  defer log.Println("Coordinator: client channels handling finished")
  
ClientLoop:
  for {
    select {
    case sock := <- cch.addclient: c.addClient(sock)
    case sock := <- cch.healthcheckRequest: c.checkHealthClient(sock)
    case sock := <- cch.readcommondata: c.readCommonData(sock)
    case sock := <- cch.runcomputation: c.runComputation(sock)
    case <- cch.collectResults: c.setGetResultsFlag()
    case <- cch.getresult: 
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

func (c *Coordinator) addWorker(w *Worker, nextIDChan chan <- uint32) {
  nextID := c.getNextWorkerID()

  w.ID = nextID
  heap.Push(&c.pool, *w)
  c.hash[nextID] = w

  log.Printf("Coordinator: added worker with ID #%v", nextID)

  nextIDChan <- nextID
}

func (c *Coordinator) getNextWorkerID() (nextID uint32) {
  nextID = c.lastWorkerID
  c.lastWorkerID++
  return nextID
}

func (c *Coordinator) addClient(sock common.Socket) {
  client := Client{
    status: common.CIdle,
    info: make(chan chan interface{}),
    ID: 0}

  defer sock.Close()

  log.Printf("Coordinator: add client with remote address %v \n", sock.RemoteAddr())
  
  canAddClient := c.client == nil
  if canAddClient {
    c.client = &client
  }
  
  client.replyInit(sock, canAddClient)
}

func (c *Coordinator) checkHealthWorker(wci WCInfo) {
  // TODO: add assert value,present = hash[addr]
  addr := wci.sock.RemoteAddr()

  log.Printf("Worker #%v healthcheck request from address: %v\n", wci.ID, addr)

  w, present := c.hash[wci.ID]
  if !present {
    log.Fatalf("Healthcheck: worker with address %v is not registered", addr)
  }

  hr := *w
  go checkHealth(hr, wci.sock, c.workerTimeout)
}

func (c *Coordinator) checkHealthClient(sock common.Socket) {
  addr := sock.RemoteAddr()

  log.Printf("Client healthcheck request from address: %v\n", addr)

  cl, present := c.client, c.client != nil
  if !present {
    log.Fatalf("Healthcheck: client with address %v is not registered", addr)
  }

  go checkHealth(cl, sock, c.clientTimeout)
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

func (c *Coordinator) sendNextTaskToWorker(wci WCInfo) {
  w, present := c.hash[wci.ID]
  if !present {
    log.Fatalf("sendNextTask: worker with id %v is not registered", wci.ID)
  }

  go w.sendNextTask(wci.sock)
}

func (c *Coordinator) workerTimeoutOccured(hr HealthReporter) {
  id := hr.GetID()
  w, present := c.hash[id]
  if !present {
    log.Fatal("Worker timeout: entity with id #%v is not present in coordinator pool", id)
  }

  if w.pending > 0 {
    // rebalance all existing tasks
    for _, task := range w.activeTasks {
      go func() { c.tasks <- *task }()
    }
  }

  delete(c.hash, id)
}

func (c *Coordinator)dispatch(task common.Task) {
  w := heap.Pop(&c.pool).(*Worker)

  // TODO: add assert pending >= 0
  
  if w.pending < w.capacity && w.pending >= 0 {
    go func(tasks chan common.Task, t common.Task) {tasks <- t} (w.tasks, task)
    w.pending++
  } else {
    // add same task again
    go func(tasks chan common.Task, t common.Task) {tasks <- t} (c.tasks, task)
  }

  heap.Push(&c.pool, w)
}

func (c *Coordinator)broadcast(task common.Task) {
  log.Println("Broadcasting task between workers")
  for _, w := range c.pool {
    w.tasks <- task
  }
}

func (c *Coordinator) setGetResultsFlag() {
  for _, w := range c.pool {
    getresults := w.GetResultsFlagChannel()
    go func(ch chan bool) {ch <- true}(getresults)
  }
}
