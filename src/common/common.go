package common

type ConnectionType byte
const (
  ClientConnection ConnectionType = iota
  WorkerConnection
)

type WorkerStatus byte
const (
  Ready WorkerStatus = iota
  Idle
  BusyAvailable
  BusyFull
  Finalized
)

type WorkerOperation byte
const (
  HealthCheck = iota
  GetTask
  TaskCompeted
  SendResult
)

type Task struct {
     id int
  // starting index and size of inner
  // multidimensional array of data
  index []int
   size []int
}
