package main

import (
	"fmt"
	"net"
	"sync"
)

type Server struct {
	IP   string
	Port int
	OnlineMap map[string]*User // online user map
	mapLock sync.RWMutex
	Message chan string // message broadcasting channel
}

// creates a new server with the given IP and port
func NewServer(ip string, port int) *Server {
	server := &Server{
		IP:   ip,
		Port: port,
		OnlineMap: make(map[string]*User),
		Message: make(chan string),
	}
	return server
}

// starts the server and listens for incoming connections
func (s *Server) Start() {
	// socket listen
	listener, err :=net.Listen("tcp", fmt.Sprintf("%s:%d", s.IP, s.Port))
	if err != nil {
		fmt.Println("Error starting server:", err)
		return
	}
	// close listen socket
	defer listener.Close()

	fmt.Printf("Server started at %s:%d\n", s.IP, s.Port)
	// start a goroutine to listen for messages on the message channel
	go s.ListenMessages()

	for {
		// accept
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Error accepting connection:", err)
			continue
		}

		// do handler
		go s.Handler(conn)
	}


}

// handles the business of the current connection
func (s *Server) Handler(conn net.Conn) {
	// the business of the current connection
	fmt.Println("the connection is established successfully")

	// create a new user
	user := NewUser(conn)

	// user online,add user to online map
	s.mapLock.Lock()
	s.OnlineMap[user.Name] = user
	s.mapLock.Unlock()

	// broadcast message to other users
	s.Broadcast(user, "已上线")

	// block the handler, otherwise the handler will exit and the connection will be closed
	select {}
}

// broadcasts a message to all online users except the sender
func (s *Server) Broadcast(user *User, msg string) {
	sendMsg := "[" + user.Name + "]:" + user.Name + ":" + msg

	s.Message <- sendMsg
}

// listens for messages on the message channel and broadcasts them to all online users
func (s *Server) ListenMessages() {
	for {
		msg := <- s.Message

		// send message to all online users
		s.mapLock.Lock()
		for _, cli := range s.OnlineMap {
			cli.C <- msg
		}
		s.mapLock.Unlock()
	}
}