package main

import (
  "../common"
)

type Coordinator struct {
  pool Pool
  done chan *Worker
}

func (c *Coordinator)dispatch(task common.Task) {
  
}

func (c *Coordinator)broadcast(task common.Task) {
  
}

func (c *Coordinator)completed(w *Worker) {
  
}
