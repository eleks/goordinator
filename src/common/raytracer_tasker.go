package common

import (
  "io"
  "math"
)

type SceneSerializer interface {
  Dump(w io.Writer) error
  Load(r io.Reader) error
  GetBinarySize() uint32
}

type Vector3 struct {
  X, Y, Z float32
}

func (v *Vector3) Dump(w io.Writer) error {
  bw := &BinWriter{W:w}
  bw.Write(v.X)
  bw.Write(v.Y)
  bw.Write(v.Z)
  return bw.Err
}

func (v *Vector3) Load(r io.Reader) error {
  br := &BinReader{R:r}
  br.Read(&v.X)
  br.Read(&v.Y)
  br.Read(&v.Z)
  return br.Err
}

func (v *Vector3) GetBinarySize() uint32 {
  return 12
}

type SphereObject struct {
  Center Vector3
  R, RSqr float64
  SurfaceColor, EmissionColor Vector3
  Transparency, Reflection float64
}

func (so *SphereObject) Dump(w io.Writer) (err error) {
  bw := &BinWriter{W:w}
  
  bw.Write(so.Center)
  bw.Write(so.R)
  bw.Write(so.RSqr)
  bw.Write(so.SurfaceColor)
  bw.Write(so.EmissionColor)
  bw.Write(so.Transparency)
  bw.Write(so.Reflection)

  return bw.Err
}

func (so *SphereObject) Load(r io.Reader) (err error) {
  br := &BinReader{R:r}
  
  br.Read(&so.Center)
  br.Read(&so.R)
  br.Read(&so.RSqr)
  br.Read(&so.SurfaceColor)
  br.Read(&so.EmissionColor)
  br.Read(&so.Transparency)
  br.Read(&so.Reflection)

  return br.Err
}

type BeginSessionTasker struct {
  Width, Height int32
  Angle int32
  SceneObjects []SceneSerializer

  ID uint32
}

func (ft *BeginSessionTasker) GetBinarySize() uint32 {
  objectsSize := uint32(0)
  for _, o := range ft.SceneObjects {
    objectsSize += o.GetBinarySize()
  }
  
  return 16 + objectsSize
}

func (ft *BeginSessionTasker) GetID() uint32 {
  return ft.ID
}

func (ft *BeginSessionTasker) GetSubTask(i uint32, gN uint32) Tasker {
  // just skip for begin session tasker
  return ft
}

func (bs *BeginSessionTasker) Dump(w io.Writer) (err error) {
  bw := &BinWriter{W:w}

  bw.Write(bs.Height)
  bw.Write(bs.Width)
  bw.Write(bs.Angle)
  bw.Write(int32(len(bs.SceneObjects)))

  if bw.Err == nil {
    for _, o := range bs.SceneObjects {
      o.Dump(w)
    }
  }

  bw.Write(bs.ID)

  return bw.Err
}

func (bs *BeginSessionTasker) Load(r io.Reader) (err error) {
  br := &BinReader{R: r}

  br.Read(&bs.Height)
  br.Read(&bs.Width)
  br.Read(&bs.Angle)

  var objectsNumber int32
  br.Read(&objectsNumber)
  
  if br.Err == nil {
    for _, o := range bs.SceneObjects {
      o.Load(r)
    }
  }

  br.Read(&bs.ID)

  return br.Err
}

type ComputeRaysTasker struct {
  IndicesCount int32
  RaysIndices []int64
  ID uint32
}

func (cr *ComputeRaysTasker) Dump(w io.Writer) (err error) {
  bw := &BinWriter{W:w}

  bw.Write(cr.IndicesCount)
  bw.Write(cr.RaysIndices)
  bw.Write(cr.ID)

  return bw.Err
}

func (cr *ComputeRaysTasker) Load(r io.Reader) (err error) {
  br := &BinReader{R:r}

  br.Read(&cr.IndicesCount)
  
  cr.RaysIndices = make([]int64, cr.IndicesCount)
  br.Read(&cr.RaysIndices)

  br.Read(&cr.ID)

  return br.Err
}

func (cr *ComputeRaysTasker) GetID() uint32 { return cr.ID }

func (cr *ComputeRaysTasker) GetSubTask(i uint32, gN uint32) Tasker {
  sliceSz := float64(cr.IndicesCount) / float64(gN)
  sliceLength := int32(math.Ceil(sliceSz))
  
  from := int32(i)*sliceLength
  to := int32(i+1)*sliceLength
  if to > cr.IndicesCount { to = cr.IndicesCount }

  return &ComputeRaysTasker{(to - from), cr.RaysIndices[from:to], 0}
}

func (cr *ComputeRaysTasker) GetBinarySize() uint32 {
  return 8 + uint32(len(cr.RaysIndices))*8
}

type RaysComputationResult struct {
  // arrays of rays indices
  // point (x,y) has index (y*imageWidth + x)
  RaysIndices []int64
  Colors []Vector3
}

func (rcm *RaysComputationResult) Dump(w io.Writer) (err error) {
  bw := &BinWriter{W:w}

  bw.Write(rcm.RaysIndices)
  bw.Write(rcm.Colors)

  return bw.Err
}

func (rcm *RaysComputationResult) Load(r io.Reader) (err error) {
  br := &BinReader{R:r}

  br.Read(&rcm.RaysIndices)
  br.Read(&rcm.Colors)

  return br.Err
}


