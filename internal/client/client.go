package client

import (
	"fmt"
	"io"
	"net"
	"os"
	"strconv"
)

type Client struct {
	ServerIP   string
	ServerPort int
	Name       string
	conn       net.Conn
	flag       int // 当前客户端选择的模式
}

func NewClient(serverIP string, serverPort int) *Client {
	client := &Client{
		ServerIP:   serverIP,
		ServerPort: serverPort,
		flag:       999,
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

func (client *Client) menu() bool {
	var flag int

	fmt.Println("1. 公聊模式")
	fmt.Println("2. 私聊模式")
	fmt.Println("3. 更新用户名")
	fmt.Println("0. 退出")

	fmt.Scanln(&flag)

	if flag >= 0 && flag <= 3 {
		client.flag = flag
		return true
	} else {
		fmt.Println("请输入合法范围内的数字")
		return false
	}
}

func (client *Client) PublicChat() {
	// 提示用户输入消息
	var chatMsg string
	fmt.Println(">>>>>>>>>请输入聊天内容，exit退出")
	fmt.Scanln(&chatMsg)

	//发给服务器
	for chatMsg != "exit" {
		// 消息不为空则发送
		if len(chatMsg) != 0 {
			sendMsg := chatMsg + "\n"
			_, err := client.conn.Write([]byte(sendMsg))
			if err != nil {
				fmt.Println("conn Write err", err)
				break
			}
		}

		chatMsg = ""
		fmt.Println(">>>>>>>>>请输入聊天内容，exit退出")
		fmt.Scanln(&chatMsg)
	}
}

// 查询在线用户
func (client *Client) SelectUsers() {
	sendMsg := "/who"
	_, err := client.conn.Write([]byte(sendMsg))
	if err != nil {
		fmt.Println("conn Write err", err)
		return
	}
}

func (client *Client) PrivateChat() {
	var remoteName string
	var chatMsg string

	client.SelectUsers()
	fmt.Println("请输入聊天对象[用户名]，exit退出")
	fmt.Scanln(&remoteName)

	if remoteName != "exit" {
		fmt.Println(">>>>>>>>>请输入消息内容")
		fmt.Scanln(&chatMsg)

		for chatMsg != "exit" {
			if len(chatMsg) != 0 {
				sendMsg := "/to "+ remoteName + " " + chatMsg + "\n\n"
				_, err := client.conn.Write([]byte(sendMsg))
				if err != nil {
					fmt.Println("conn Write err", err)
					break
				}
			}

			chatMsg = ""
			fmt.Println(">>>>>>>>>请输入聊天内容，exit退出")
			fmt.Scanln(&chatMsg)
		}

		// 退出，重新进入到选择聊天对象部分
		client.SelectUsers()
		fmt.Println("请输入聊天对象[用户名]，exit退出")
		fmt.Scanln(&remoteName)
	}
}

func (client *Client) UpdateName() bool {
	fmt.Println(">>>>>>>>请输入用户名")
	fmt.Scanln(&client.Name)

	sendMsg := "/rename " + client.Name + "\n"
	_, err := client.conn.Write([]byte(sendMsg))
	if err != nil {
		fmt.Println("conn.Write err:", err)
		return false
	}
	return true
}

// 处理server返回的消息，直接显示到标准输出即可
func (client *Client) DealResponse() {
	// 一旦client.conn有数据，就直接copy到stdout标准输出，永久阻塞监听
	io.Copy(os.Stdout, client.conn)
}

func (client *Client) Run() {
	for {
		if !client.menu() {
			continue
		}

		switch client.flag {
		case 0:
			return
		case 1:
			client.PublicChat()
		case 2:
			client.PrivateChat()
		case 3:
			client.UpdateName()
		}
	}
}
