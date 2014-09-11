package main

import (
  "../common"
)

type Tasker interface {
  ExportData() (common.GenericData, error)
  ImportData(data common.GenericData) error
  GrindIntoSubtasks(n int) ([]Tasker, error
  MergeSubtasks(subtasks []Tasker) error
}
