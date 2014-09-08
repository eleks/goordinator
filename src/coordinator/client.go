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
  rmclient chan *Client
}

type Client struct {
  sock common.Socket
  commondata common.DataArray
  // buffered channel operates with int32
  info chan chan interface{}

  tasksCount uint32
  doneTasksCount uint32

  status common.ClientStatus
}

func (c *Client) CloseSock() { c.sock.Close() }
func (c *Client) GetSock() common.Socket { return c.sock }

func (c *Client) GetStatus() interface{} { return c.status }
func (c *Client) GetStatusChannel() chan chan interface{} { return c.info }

func (c *Client) SetHealthStatus(status byte) { c.status = common.ClientStatus(status) }
func (c *Client) GetHealthReply() interface{} { return uint32(c.done_tasks_count * 100 / c.tasks_count) }


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
