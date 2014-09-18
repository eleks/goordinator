package common

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
  Data []float32  

  // dim1 * dim2 - size of 2d array
  // dim3 - size of array of 2d arrays
  Dim1, Dim2, Dim3 int

  ID int
}

func (ft TaskParameterFloat) Dump(w io.Writer) (err error) {
  err = binary.Write(w, binary.BigEndian, ft.GetSize())
  if err != nil {
    return err
  }

  err = binary.Write(w, binary.BigEndian, ft.Dim1)
  if err != nil {
    return err
  }

  err = binary.Write(w, binary.BigEndian, ft.Dim2)
  if err != nil {
    return err
  }

  err = binary.Write(w, binary.BigEndian, ft.Dim3)
  if err != nil {
    return err
  }

  err = binary.Write(w, binary.BigEndian, ft.Data)
  if err != nil {
    return err
  }

  return nil
}

func (ft TaskParameterFloat) GetSize() uint32 {
  sizeofFloat32 := 4
  dimensionsSize := uint32(3*sizeofFloat32)
  dataSize := uint32(ft.Dim1 * ft.Dim2 * ft.Dim3 * sizeofFloat32)
  return dimensionsSize + dataSize
}

func (ft TaskParameterFloat) GetID() int { return ft.ID }

func (ft *TaskParameterFloat) Get(i, j, k int) float32 {
  return ft.Data[(i*ft.Dim2 + j) + ft.Dim1 * ft.Dim2 * k]
}

func (ft *TaskParameterFloat) Get2D(i, j int) float32 {
  return ft.Data[i*ft.Dim2 + j]
}

type ExecFunc func(float32, float32) float32

func (ft *TaskParameterFloat) Exec(f ExecFunc, param float32) {
  for i, v := range ft.Data {
    ft.Data[i] = f(v, param)
  }
}

func (ft TaskParameterFloat) GobEncode() ([]byte, error) {
  w := new(bytes.Buffer)
  encoder := gob.NewEncoder(w)

  // TODO: handle errors here
  err := encoder.Encode(ft.Dim1)
  err = encoder.Encode(ft.Dim2)
  err = encoder.Encode(ft.Dim3)

  err = encoder.Encode(ft.Data)

  return w.Bytes(), err
}

func (ft TaskParameterFloat) GobDecode(buf []byte) error {
  r := bytes.NewBuffer(buf)
  decoder := gob.NewDecoder(r)

  err := decoder.Decode(&ft.Dim1)
  err = decoder.Decode(&ft.Dim2)
  err = decoder.Decode(&ft.Dim3)
  
  ft.Data = make([]float32, ft.Dim1 * ft.Dim2 * ft.Dim3)
  err = decoder.Decode(&ft.Data)
  return err
}

func (ft TaskParameterFloat) GetSubTask(i int, grindNumber int) Tasker {
  h, w, d := ft.Dim1/grindNumber, ft.Dim2, ft.Dim3
  var t Tasker
  t = TaskParameterFloat{ft.Data[i*w:(i+1)*w], h, w, d, 0}
  return t
}

func (ft TaskParameterFloat) GrindIntoSubtasks(n int) (tasks []Tasker, err error) {
  if ft.Dim1 % n != 0 {
    return nil, errors.New("First dimension should be divisible by n without remain")
  }

  h := ft.Dim1 / n
  tasks = make([]Tasker, n)
  
  w, d := ft.Dim2, ft.Dim3
  
  for i := range tasks {
    tasks[i] = TaskParameterFloat{ft.Data[i*w:(i+1)*w], h, w, d, 0}
  }

  return tasks, nil
}


