package server

import (
	"fmt"
	"io"
	"net"
	"strings"
	"sync"
	"time"
)

type Server struct {
	IP        string
	Port      int
	OnlineMap map[string]*User // online user map
	mapLock   sync.RWMutex
	Message   chan string // message broadcasting channel
}

// creates a new server with the given IP and port
func NewServer(ip string, port int) *Server {
	server := &Server{
		IP:        ip,
		Port:      port,
		OnlineMap: make(map[string]*User),
		Message:   make(chan string),
	}
	return server
}

// starts the server and listens for incoming connections
func (s *Server) Start() {
	// socket listen
	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", s.IP, s.Port))
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
	user := NewUser(conn, s)

	user.Online()

	// channel to detect if the user is still alive
	isLive := make(chan bool)

	// listen for messages from the user
	go func() {
		buf := make([]byte, 4096)
		for {
			n, err := conn.Read(buf)
			if n == 0 {
				user.Offline()
				return
			}

			if err != nil && err != io.EOF {
				fmt.Println("Error reading from connection:", err)
				return
			}

			// extract the message from the buffer and broadcast it to other users
			msg := strings.TrimSpace(string(buf[:n]))
			user.DoMessage(msg)

			isLive <- true // user is still alive
		}
	}()

	for {
		select {
		case <-isLive:
			// do nothing, to reset the timer
		case <-time.After(300 * time.Second):
			// already offline
			conn.Write([]byte("你被踢了\n"))

			// close
			close(user.C)
			conn.Close()

			return
		}
	}
}

// broadcasts a message to all online users
func (s *Server) Broadcast(user *User, msg string) {
	sendMsg := "[" + user.Addr + "]:" + user.Name + ":" + msg

	s.Message <- sendMsg
}

// listens for messages on the message channel and broadcasts them to all online users
func (s *Server) ListenMessages() {
	for {
		msg := <-s.Message

		// send message to all online users
		s.mapLock.Lock()
		for _, cli := range s.OnlineMap {
			cli.C <- msg
		}
		s.mapLock.Unlock()
	}
}
