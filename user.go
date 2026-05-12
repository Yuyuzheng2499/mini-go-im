package main

import (
	"net"
)

type User struct {
	Name string
	Addr string
	C chan string
	conn net.Conn
}

// creates a new user with the given connection
func NewUser(conn net.Conn) *User {
	userAddr := conn.RemoteAddr().String()

	user := &User {
		Name: userAddr,
		Addr: userAddr,
		C: make(chan string),
		conn: conn,
	}
	// start a goroutine to listen for messages on the user's channel
	go user.ListenMessages()

	return user
}

// listens for messages on the user's channel and sends them to the user's connection
func (u *User) ListenMessages() {
	for {
		msg := <- u.C
		u.conn.Write([]byte(msg + "\n"))
	}
}