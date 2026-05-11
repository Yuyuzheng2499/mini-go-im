package main

import (
	"fmt"
	"net"
)

type Server struct {
	IP   string
	Port int
}

// creates a new server with the given IP and port
func NewServer(ip string, port int) *Server {
	server := &Server{
		IP:   ip,
		Port: port,
	}
	return server
}

func (s *Server) Handler(conn net.Conn) {
	// the business of the current connection
	fmt.Println("the connection is established successfully")
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
