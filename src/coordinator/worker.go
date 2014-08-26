package main

import (
  "../common"
  "net"
  "time"
  "fmt"
  "encoding/binary"
)

type WorkerInfo struct {
  pending uint32
  status common.WorkerStatus
}

type Worker struct {
  index uint32
  sock common.Socket
  tasks chan common.Task
  // buffered channel
  info chan WorkerInfo
  // buffered channel
  getInfo chan bool
  // number of pending tasks
  pending uint32
  capacity uint32
  status common.WorkerStatus
}

func (w *Worker) checkHealth(timeout chan *Worker) {
  defer sock.Close()
  
  healthcheck := make(chan common.WorkerStatus, 1)
  reply := make(chan uint32)
  done := make(chan bool, 1)
  
  go func() {
    worker_status_buf := make([]byte, 1)
    tasks_available_buf := make([]byte, 4)

  healthLoop:
    for {      
      _, err := io.ReadFull(w.sock, worker_status_buf)
      if err != nil {
        // TODO: send errors to channel
        fmt.Printf(err)
        break healthLoop
      }

      worker_status := common.WorkerStatus(worker_status_buf[0])
      healthcheck <- worker_status

      select {
      case tasks_available := <- reply: {
        err := binary.Write(w.sock, binary.BigEndian, tasks_available)
        if err != nil {
          // TODO: send errors to channel
          fmt.Printf(err)
          break healthLoop
        }
      }
      case <- done: break healthLoop
      }
    }
  }()
  
Loop:
  for {
    select {
    case status := <- healthcheck: {
      w.status = status
      reply <- w.pending
    }
    case <- w.getInfo: w.info <- WorkerInfo {w.pending, w.status}
    case <- time.After(1 * time.Second): {      
      // TODO: change timeout
      done <- true
      timeout <- w
      break Loop
    }
    }
  }
}

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


