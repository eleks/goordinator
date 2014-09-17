package main

import (
  "bytes"
  "encoding/binary"
  "encoding/gob"
  "errors"
  "io"
)

// implements Tasker
type TaskParameterFloat struct {
  // to allow efficient memory caching
  data []float32  

  // dim1 * dim2 - size of 2d array
  // dim3 - size of array of 2d arrays
  dim1, dim2, dim3 int

  ID int
}

func (ft TaskParameterFloat) Dump(w io.Writer) (err error) {
  err = binary.Write(w, binary.BigEndian, ft.GetSize())
  if err != nil {
    return err
  }

  err = binary.Write(w, binary.BigEndian, ft.dim1)
  if err != nil {
    return err
  }

  err = binary.Write(w, binary.BigEndian, ft.dim2)
  if err != nil {
    return err
  }

  err = binary.Write(w, binary.BigEndian, ft.dim3)
  if err != nil {
    return err
  }

  err = binary.Write(w, binary.BigEndian, ft.data)
  if err != nil {
    return err
  }

  return nil
}

func (ft TaskParameterFloat) GetSize() uint32 {
  sizeofFloat32 := 4
  dimensionsSize := uint32(3*sizeofFloat32)
  dataSize := uint32(ft.dim1 * ft.dim2 * ft.dim3 * sizeofFloat32)
  return dimensionsSize + dataSize
}

func (ft TaskParameterFloat) GetID() int { return ft.ID }

func (ft *TaskParameterFloat) Get(i, j, k int) float32 {
  return ft.data[(i*ft.dim2 + j) + ft.dim1 * ft.dim2 * k]
}

func (ft *TaskParameterFloat) Get2D(i, j int) float32 {
  return ft.data[i*ft.dim2 + j]
}

func (ft TaskParameterFloat) GobEncode() ([]byte, error) {
  w := new(bytes.Buffer)
  encoder := gob.NewEncoder(w)

  // TODO: handle errors here
  err := encoder.Encode(ft.dim1)
  err = encoder.Encode(ft.dim2)
  err = encoder.Encode(ft.dim3)

  err = encoder.Encode(ft.data)

  return w.Bytes(), err
}

func (ft TaskParameterFloat) GobDecode(buf []byte) error {
  r := bytes.NewBuffer(buf)
  decoder := gob.NewDecoder(r)

  err := decoder.Decode(&ft.dim1)
  err = decoder.Decode(&ft.dim2)
  err = decoder.Decode(&ft.dim3)
  
  ft.data = make([]float32, ft.dim1 * ft.dim2 * ft.dim3)
  err = decoder.Decode(&ft.data)
  return err
}

func (ft TaskParameterFloat) GetSubTask(i int, grindNumber int) Tasker {
  h, w, d := ft.dim1/grindNumber, ft.dim2, ft.dim3
  var t Tasker
  t = TaskParameterFloat{ft.data[i*w:(i+1)*w], h, w, d, 0}
  return t
}

func (ft TaskParameterFloat) GrindIntoSubtasks(n int) (tasks []Tasker, err error) {
  if ft.dim1 % n != 0 {
    return nil, errors.New("First dimension should be divisible by n without remain")
  }

  h := ft.dim1 / n
  tasks = make([]Tasker, n)
  
  w, d := ft.dim2, ft.dim3
  
  for i := range tasks {
    tasks[i] = TaskParameterFloat{ft.data[i*w:(i+1)*w], h, w, d, 0}
  }

  return tasks, nil
}


