package main

import (
  "../common"
)

type WorkerInfo struct {
  pending uint32
  status common.WorkerStatus
}

type WorkerChannels struct {
  addworker chan *Worker
  healthcheck_request chan common.Socket
  rmworker chan *Worker
}

type Worker struct {
  index int
  sock common.Socket
  tasks chan common.Task
  // buffered channel operates with common.WorkerInfo
  info chan interface{}
  // buffered channel
  getInfo chan bool
  // number of pending tasks
  pending uint32
  capacity uint32
  status common.WorkerStatus
}

func (w *Worker) CloseSock() { w.sock.Close() }
func (w *Worker) GetSock() common.Socket { return w.sock }

func (w *Worker) GetStatusRequestChannel() chan bool { return w.getInfo }
func (w *Worker) GetStatus() interface{} { return w.status }
func (w *Worker) GetStatusChannel() chan interface{} { return w.info }
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


