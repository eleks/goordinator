package main

import (
  "../common"
  "log"
  "errors"
)

type MatrixComputator struct {
  factor float32
}

func (mc MatrixComputator) beginSession(task common.Task) error {
  if len(task.Parameters) != 1 {
    return errors.New("Invalid parameters")
  }
  
  gd := task.Parameters[0]
  tpf := new(common.TaskParameterFloat)
  err := tpf.GobDecode(gd.Data)

  if err != nil {
    return err
  }

  if tpf.GetSize() != 1 {
    return errors.New("Invalid length of initial parameter")
  }

  mc.factor = tpf.Get(0, 0, 0)
  log.Printf("Received common parameter %v\n", mc.factor)
  return nil
}

func (mc MatrixComputator) computeTask(task common.Task) (cr *common.ComputationResult, err error) {
  cr = new(common.ComputationResult)

  if len(task.Parameters) != 1 {
    return cr, errors.New("Invalid parameters")
  }

  gd := task.Parameters[0]

  tpf := new(common.TaskParameterFloat)
  err = tpf.GobDecode(gd.Data)

  if err != nil {
    return cr, err
  }

  if tpf.Dim3 != 0 {
    return cr, errors.New("Invalid parameters size")
  }

  var f common.ExecFunc
  f = func(v float32, factor float32) float32 { return v * factor }
  tpf.Exec(f, mc.factor)

  var replyBuf []byte
  replyBuf, err = tpf.GobEncode()

  if err != nil {
    return cr, err
  }

  cr.Data = replyBuf
  cr.Size = uint32(len(replyBuf))

  return cr, nil
}

func (mc MatrixComputator) endSession() error {
  return nil
}
