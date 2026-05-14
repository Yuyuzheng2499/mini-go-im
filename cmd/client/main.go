package main

import (
	"fmt"

	imclient "github.com/Yuyuzheng2499/mini-go-im/internal/client"
)

func main() {
	client := imclient.NewClient("127.0.0.1", 8888)
	if client == nil {
		fmt.Println(">>>>>>>链接服务器失败...")
		return
	}
	fmt.Println(">>>>>>>链接服务器成功...")

	select {}
}
