package main

import (
  "../common"
  "encoding/binary"
  "log"
)

type WorkerChannels struct {
  addworker chan *Worker
  initWorker chan *Worker
  healthcheckRequest chan WCInfo
  nextID chan uint32
  gettaskRequest chan WCInfo
  taskresult chan common.Socket
  rmworker chan *Worker
}

// worker connection info
type WCInfo struct {
  ID uint32
  sock common.Socket
}

type Worker struct {
  index int
  // buffered channel (buffer size is capacity)
  tasks chan common.Task
  getResults chan bool
  // buffered channel operates with common.WorkerInfo
  cinfo chan interface{}
  ccinfo chan chan interface{}
  updatePending chan int32
  activeTasks map[int64]*common.Task
  // if positive means number of pending tasks 
  // else means number of task to retrieve from worker
  // used only from coordinator's select
  pending int32
  // used only for healthcheck
  cachedPending int32
  capacity int32
  tasksDone int32
  ID uint32
  getresultsFlag bool
}

func (w *Worker) IncPendingTasks() {
  w.pending++
  w.updatePending <- w.pending
}

func (w Worker) GetStatus() interface{} { return w.tasksDone }
func (w Worker) GetStatusChannel() chan chan interface{} { return w.ccinfo }

func (w Worker) SetHealthStatus(tasksDone int32) { w.tasksDone = tasksDone }

func (w Worker) GetHealthReply() interface{} {
  var result int32
  if !w.getresultsFlag {
    result = w.cachedPending
  } else {
    result = -1
  }
  
  return result
}

func (w Worker) SetHealthReply(pending int32) {
  w.cachedPending = pending
}

func (w Worker) GetID() uint32 { return w.ID }

func (w Worker) GetResultsFlagChannel() chan bool { return w.getResults }
func (w Worker) SetGetResultsFlag() { w.getresultsFlag = true }

func (w Worker) GetUpdateReplyChannel() chan int32 { return w.updatePending }

type Pool []*Worker

func (p Pool) Len() int { return len(p) }

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

func (w *Worker) RetrieveStatus() uint32 {
  w.ccinfo <- w.cinfo
  doneTasksCount := <- w.cinfo
  return doneTasksCount.(uint32)
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
