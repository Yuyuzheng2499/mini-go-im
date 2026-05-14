package main

import imserver "github.com/Yuyuzheng2499/mini-go-im/internal/server"

func main() {
	server := imserver.NewServer("127.0.0.1", 8888)
	server.Start()
}
