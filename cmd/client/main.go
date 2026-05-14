package main

import (
	"flag"
	"fmt"

	imclient "github.com/Yuyuzheng2499/mini-go-im/internal/client"
)

func main() {
	serverIP := flag.String("ip", "127.0.0.1", "设置服务器IP地址(默认是127.0.0.1)")
	serverPort := flag.Int("port", 8888, "设置服务器端口号(默认8888)")
	flag.Parse()

	client := imclient.NewClient(*serverIP, *serverPort)
	if client == nil {
		fmt.Println(">>>>>>>链接服务器失败...")
		return
	}
	fmt.Println(">>>>>>>链接服务器成功...")

	select {}
}
