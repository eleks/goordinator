package main

import (
  "../common"
  "bytes"
  "encoding/gob"
  "errors"
)

// implements Tasker
type TaskParameterFloat struct {
    // to allow efficient memory caching
  data []float32  

  // dim1 * dim2 - size of 2d array
  // dim3 - size of array of 2d arrays
  dim1, dim2, dim3 int
}

func (ft *TaskParameterFloat) Get(i, j, k int) float32 {
  return ft.data[(i*ft.dim2 + j) + ft.dim1 * ft.dim2 * k]
}

func (ft *TaskParameterFloat) Get2D(i, j int) float32 {
  return ft.data[i*ft.dim2 + j]
}

func (ft *TaskParameterFloat) GobEncode() ([]byte, error) {
  w := new(bytes.Buffer)
  encoder := gob.NewEncoder(w)

  // TODO: handle errors here
  err := encoder.Encode(ft.dim1)
  err = encoder.Encode(ft.dim2)
  err = encoder.Encode(ft.dim3)

  err = encoder.Encode(ft.data)

  return w.Bytes(), nil
}

func (ft *TaskParameterFloat) GobDecode(buf []byte) error {
  r := bytes.NewBuffer(buf)
  decoder := gob.NewDecoder(r)

  err := decoder.Decode(&ft.dim1)
  err = decoder.Decode(&ft.dim2)
  err = decoder.Decode(&ft.dim3)
  
  ft.array = make([]float32, ft.dim1 * ft.dim2 * ft.dim3)
  err = decoder.Decode(&ft.data)
}

func (ft *TaskParameterFloat) GrindIntoSubtasks(n int) (tasks []TaskParameterFloat, err error) {
  if ft.dim1 % n != 0 {
    return nil, errors.New("First dimension should be divisible by n without remain")
  }

  h := ft.dim1 / n
  tasks = make([]TaskParameterFloat, n)
  
  w, d := ft.dim2, ft.dim3
  
  for i := range tasks {
    tasks[i] := TaskParameterFloat{ft.data[i*w:(i+i)*w], h, w, d}
  }

  return tasks, nil
}


