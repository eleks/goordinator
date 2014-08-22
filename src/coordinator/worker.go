package main

import (
  "../common"
  "net"
)

type Worker struct {
  index int
  addr net.Addr
  tasks chan common.Task
  // number of pending tasks
  pending int
  capacity int
  status common.WorkerStatus
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


