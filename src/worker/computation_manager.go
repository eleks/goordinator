package main

import (
  "../common"
  "net"
  "log"
  "encoding/binary"
)

type ComputationManager struct {
  healthcheck_response chan int
  status_info chan chan common.WorkerStatus
  pending_tasks_count int
  tasks chan common.Task
  tasks_results map[int]common.ComputationResult
  results chan common.ComputationResult
  status common.WorkerStatus
  stop_messages chan chan error
  // buffered
  stop_computations chan bool
}

func (cm *ComputationManager) handleCommands() {
  var err error
Loop:
  for {
    select {
    case info_channel := <- cm.status_info: info_channel <- cm.status
    case pending := <- cm.healthcheck_response: {
      if pending > cm.pending_tasks_count {
        go cm.downloadNewTask()
        cm.pending_tasks_count++
      } else if pending < 0 {
        // TODO: rewrite task_results to list
        go cm.sendTaskResult()
      }
    }
    case result := <- cm.results: {
      cm.tasks_results[result.ID] = result
    }
    case error_channel := <- cm.stop_messages: {
      error_channel <- err
      break Loop
    }
    }
  }
}

func (cm *ComputationManager) downloadNewTask() {
  conn, _ := net.Dial("tcp", *caddr)

  binary.Write(conn, binary.BigEndian, common.WGetTask)

  var taskID int
  binary.Read(conn, binary.BigEndian, &taskID)

  log.Printf("Downloading task #%v parameters", taskID)
  
  dataArray, _, _ := common.ReadDataArray(conn)
  // TODO: handle error here

  log.Printf("Received task #%v", taskID)
  cm.tasks <- common.Task{taskID, dataArray}
}

func (cm *ComputationManager) computeTasks() {
Loop:
  for {
    select {
    case task := <- cm.tasks: {
      log.Printf("Task #%v computation started", task.ID)

      // TODO: add computation itself
      
      log.Printf("Task #%v computation finished", task.ID)
      var results common.ComputationResult
      cm.results <- results
    }
    case <- cm.stop_computations:
      break Loop
    }
  }
}

func (cm *ComputationManager) sendTaskResult(res common.ComputationResult) {
  conn, _ := net.Dial("tcp", *caddr)

  binary.Write(conn, binary.BigEndian, common.WSendResult)

  binary.Write(conn, binary.BigEndian, res.ID)
  
  common.WriteDataArray(conn, res)
}

func (cm *ComputationManager) stop() error {
  cm.stop_computations <- true
  errors := make(chan error)
  cm.stop_messages <- errors
  return <- errors
}
