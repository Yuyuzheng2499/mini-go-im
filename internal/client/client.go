package client

import (
	"fmt"
	"net"
	"strconv"
)

type Client struct {
	ServerIP   string
	ServerPort int
	Name       string
	conn       net.Conn
}

func NewClient(serverIP string, serverPort int) *Client {
	client := &Client{
		ServerIP:   serverIP,
		ServerPort: serverPort,
	}

	// conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", serverIP, serverPort))
	conn, err := net.Dial("tcp", net.JoinHostPort(serverIP, strconv.Itoa(serverPort)))

	if err != nil {
		fmt.Println("net.Dial error:", err)
		return nil
	}
	client.conn = conn

	return client
}
