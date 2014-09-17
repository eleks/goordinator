package main

import (
  "../common"
  "log"
)

type ClientChannels struct {
  addclient chan *Client
  healthcheckRequest chan common.Socket
  readcommondata chan common.Socket
  runcomputation chan common.Socket
  collectResults chan bool
  getresult chan common.Socket
  computationResults chan common.ComputationResult
  rmclient chan *Client
}

type Client struct {
  commondata common.DataArray
  // buffered channel operates with int32
  info chan chan interface{}

  tasksCount uint32
  doneTasksCount uint32
  status common.ClientStatus

  ID uint32
}

func (c *Client) GetStatus() interface{} { return c.status }
func (c *Client) GetStatusChannel() chan chan interface{} { return c.info }

func (c *Client) SetHealthStatus(status byte) { c.status = common.ClientStatus(status) }
func (c *Client) GetHealthReply() interface{} { return uint32(c.doneTasksCount * 100 / c.tasksCount) }

func (w *Client) GetResultsFlagChannel() chan bool { return make(chan bool) }
func (w *Client) SetGetResultsFlag() { }

func (c *Client) replyInit(success bool) {
  log.Printf("Replying to client. Connection was successful: %v\n", success)
  
  var status byte
  if success {
    status = 1
  } else {
    status = 0
  }

  binary.Write(c.sock, binary.BigEndian, status)
}

func (c *Client) RetrieveStatus() common.ClientStatus {
  hchannel := make(chan interface{})
  c.info <- hchannel
  status := <- hchannel
  return status.(common.ClientStatus)
}
