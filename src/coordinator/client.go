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
  getresults chan bool
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
  
  statusBuf := make([]byte, 1)
  if success {
    statusBuf[0] = 1
  } else {
    statusBuf[0] = 0
  }

  c.sock.Write(statusBuf)
}

func (c *Client) RetrieveStatus() common.ClientStatus {
  hchannel := make(chan interface{})
  c.info <- hchannel
  status := <- hchannel
  return status.(common.ClientStatus)
}
