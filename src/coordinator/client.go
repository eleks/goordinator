package main

import (
  "../common"
)

type ClientChannels struct {
  addclient chan *Client
  healthcheck_request chan common.Socket
  readcommondata chan common.Socket
  rmclient chan *Client
}

type Client struct {
  sock common.Socket
  common_data []common.Parameter
  status common.ClientStatus
}

func (c *Client) CloseSock() { c.sock.Close() }
func (c *Client) GetSock() common.Socket { return c.sock }

func (c *Client) GetStatusRequestChannel() chan bool { return make(chan bool) }
func (c *Client) GetStatus() interface{} { return c.status }
func (c *Client) GetStatusChannel() chan interface{} { return make(chan interface {}) }
func (c *Client) SetHealthStatus(status byte) { c.status = common.ClientStatus(status) }
func (c *Client) GetHealthReply() interface{} { return 1 }


func (c *Client) replyInit(success bool) {
  defer c.sock.Close()

  connection_status_buf := make([]byte, 1)
  if success {
    connection_status_buf[0] = 1
  } else {
    connection_status_buf[0] = 0
  }

  c.sock.Write(connection_status_buf)
}
