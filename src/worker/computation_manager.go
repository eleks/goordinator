package main

import (
  "../common"
  "net"
  "log"
  "encoding/binary"
)

type ComputationManager struct {
  // always greater than zero (0 == unassigned)
  ID uint32
  healthcheckResponse chan int
  statusInfo chan chan uint32
  pendingTasksCount int
  tasks chan common.Task
  results map[int]common.ComputationResult
  chResults chan common.ComputationResult
  stopMessages chan chan error
  tasksDone int  
  sendingMode bool
  // buffered
  stopComputations chan bool
}

func (cm *ComputationManager) handleCommands() {
  var err error
LoopCommands:
  for {
    select {
    case infoChannel := <- cm.statusInfo: infoChannel <- uint32(cm.tasksDone)
    case pending := <- cm.healthcheckResponse: {
      if pending > cm.pendingTasksCount {
        go cm.downloadNewTask()
        cm.pendingTasksCount++
      } else if pending < 0 && !cm.sendingMode {
        cm.sendingMode = true
        cm.sendTaskResults()
      }
    }
    case result := <- cm.chResults: {
      if !cm.sendingMode {
        cm.results[result.ID] = result
        cm.tasksDone = len(cm.results)
      } else {
        log.Fatal("Attempt to save result while sending results")
      }
    }
    case errorChannel := <- cm.stopMessages: {
      errorChannel <- err
      break LoopCommands
    }
    }
  }
}

func (cm *ComputationManager) downloadNewTask() {
  conn, _ := net.Dial("tcp", *caddr)

  binary.Write(conn, binary.BigEndian, common.WGetTask)
  binary.Write(conn, binary.BigEndian, cm.ID)

  var taskID int64
  binary.Read(conn, binary.BigEndian, &taskID)

  log.Printf("Downloading task #%v parameters", taskID)
  
  dataArray, _, _ := common.ReadDataArray(conn)
  // TODO: handle error here

  log.Printf("Received task #%v", taskID)
  cm.tasks <- common.Task{taskID, dataArray}
}

func (cm *ComputationManager) processTasks() {
  isFirst := true
  
Loop:
  for {
    select {
    case task := <- cm.tasks: {
      if isFirst {
        // TODO: handle common params
      } else {      
            log.Printf("Task #%v computation started", task.ID)

        // TODO: add computation itself
        // TODO: add failed computations too for overall count
      
        log.Printf("Task #%v computation finished", task.ID)
        var cr common.ComputationResult
        cm.ch_results <- cr
      }
    }
      // for healthcheck fail
    case <- cm.stopComputations:
      break Loop
    }
  }
}

func (cm *ComputationManager) sendTaskResults() {
  log.Printf("Sending task results to coordinator.. Launching goroutines")
  rcount := len(cm.results)
  waitAll := make([]chan bool, rcount)
  for i := range waitAll { waitAll[i] := make(chan bool) }
  
  for key, cr := range cm.results {
    go sendTaskResult(cr, waitAll[i])
    delete(cm.results, key)    
  }

  log.Printf("Waiting for all send task completed")

  // wait all and stop coordinator
  go func(waitArr[] chan bool, c *ComputationManager) {
    for i := range waitArr { <- waitArr[i] }
    c.stop()
  }(waitAll)
}

func sendTaskResult(cr common.ComputationResult, finished chan bool) {
  conn, _ := net.Dial("tcp", *caddr)

  binary.Write(conn, binary.BigEndian, common.WSendResult)

  binary.Write(conn, binary.BigEndian, cr.ID)

  common.WriteGenericData(conn, cr)

  finished <- true
}

func (cm *ComputationManager) stop() error {
  log.Printf("ComputationManager: stop received")
  cm.stopComputations <- true
  errors := make(chan error)
  cm.stopMessages <- errors
  return <- errors
}
