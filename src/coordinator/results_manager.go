package main

import (
  "../common"
  "encoding/binary"
)

func handleWorkerGetResults(tasksresults <-chan common.Socket, computationResults chan<- common.ComputationResult) {
  var taskID int64
  var err error

  for sock := range tasksresults {
    err = binary.Read(sock, binary.BigEndian, &taskID)

    if err == nil {
      gd, err = common.ReadGenericData(sock)

      if err == nil {
        go func(crch chan-> commmon.ComputationResult, ch common.ComputationResult) {crch <- ch} (computationResults, ComputationResult{taskID, gd})
      }
    }

    sock.Close()
  }
}

func handleClientGetResults(getResult <-chan common.Socket, computationResults <-chan common.ComputationResult) {
  var err error
  for sock := range getResult {
    cr := <- computationResults

    err = binary.Write(sock, binary.BigEndian, cr.ID)
    if err == nil {
      common.WriteGenericData(sock, cr)
    }

    sock.Close()
  }
}
