package main

import (
  "../common"
  "log"
)

type ClientChannels struct {
  addclient chan *Client
  healthcheck_request chan common.Socket
  readcommondata chan common.Socket
  runcomputation chan common.Socket
  rmclient chan *Client
}

type Client struct {
  sock common.Socket
  commondata common.DataArray
  // buffered channel operates with int32
  info chan interface{}
  // buffered channel
  getInfo chan bool

  tasks_count uint32
  done_tasks_count uint32

  status common.ClientStatus
}

func (c *Client) CloseSock() { c.sock.Close() }
func (c *Client) GetSock() common.Socket { return c.sock }

func (c *Client) GetStatusRequestChannel() chan bool { return c.getInfo }
func (c *Client) GetStatus() interface{} { return c.status }
func (c *Client) GetStatusChannel() chan interface{} { return c.info }
func (c *Client) SetHealthStatus(status byte) { c.status = common.ClientStatus(status) }
func (c *Client) GetHealthReply() interface{} { return uint32(c.done_tasks_count * 100 / c.tasks_count) }


func (c *Client) replyInit(success bool) {
  log.Printf("Replying to client. Connection was successful: %v\n", success)
  
  connection_status_buf := make([]byte, 1)
  if success {
    connection_status_buf[0] = 1
  } else {
    connection_status_buf[0] = 0
  }

  c.sock.Write(connection_status_buf)
}
