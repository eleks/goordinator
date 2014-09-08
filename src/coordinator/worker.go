package main

import (
  "../common"
  "encoding/binary"
  "log"
)

type WorkerChannels struct {
  addworker chan *Worker
  healthcheckRequest chan common.Socket
  gettaskRequest chan common.Socket
  rmworker chan *Worker
}

type Worker struct {
  index int
  sock common.Socket
  // buffered channel (buffer size is capacity)
  tasks chan common.Task
  stop chan bool
  // buffered channel operates with common.WorkerInfo
  cinfo chan interface{}
  ccinfo chan chan interface{}
  activeTasks map[int]*common.Task
  // if positive means number of pending tasks 
  // else means number of task to retrieve from worker
  pending int
  capacity int
  status common.WorkerStatus
}

func (w *Worker) CloseSock() { w.sock.Close() }
func (w *Worker) GetSock() common.Socket { return w.sock }

func (w *Worker) GetStatus() interface{} { return w.status }
func (w *Worker) GetStatusChannel() chan chan interface{} { return w.ccinfo }

func (w *Worker) SetHealthStatus(status byte) { w.status = common.WorkerStatus(status) }
func (w *Worker) GetHealthReply() interface{} { return w.pending }


type Pool []*Worker

func (p Pool) Len() int { return len(p); }

func (p Pool) Less(i, j int) bool {
  return p[i].pending < p[j].pending
}

func (p *Pool) Swap(i, j int) {
  a := *p
  a[i], a[j] = a[j], a[i]
  a[i].index = i
  a[j].index = j
}

func (p *Pool) Push(x interface{}) {
  a := *p
  n := len(a)
  a = a[0:n+1]
  w := x.(*Worker)
  a[n] = w
  w.index = n
  *p = a
}

func (p *Pool) Pop() interface{} {
  a := *p
  *p = a[0 : len(a) - 1]
  w := a[len(a) - 1]
  w.index = -1
  return w
}

func (w *Worker) RetrieveStatus() common.WorkerStatus {
  w.ccinfo <- w.cinfo
  status := <- w.cinfo
  return status.(common.WorkerStatus)
}

func (w *Worker) doWork() {
  log.Printf("Worker with index %v started working", w.index)
Loop:
  for {
    select {
    case <- w.stop:
      break Loop
    }
  }
}

func (w *Worker) sendNextTask(sock common.Socket) error {
  task := <- w.tasks

  _, ok := w.activeTasks[task.ID]
  if !ok {
    w.activeTasks[task.ID] = &task
  } else {
    log.Fatalf("Task with ID %v is already sent to this worker", task.ID)
  }

  err := binary.Write(sock, binary.BigEndian, task.ID)
  // TODO: handle error

  if err == nil {
    err = common.WriteDataArray(sock, task.Parameters)
  }

  return err
}
