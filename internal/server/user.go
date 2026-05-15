package server

import (
	"net"
	"strings"
)

type User struct {
	Name   string
	Addr   string
	C      chan string
	conn   net.Conn
	server *Server
}

// creates a new user with the given connection
func NewUser(conn net.Conn, server *Server) *User {
	userAddr := conn.RemoteAddr().String()

	user := &User{
		Name:   userAddr,
		Addr:   userAddr,
		C:      make(chan string),
		conn:   conn,
		server: server,
	}
	// start a goroutine to listen for messages on the user's channel
	go user.ListenMessages()

	return user
}

// listens for messages on the user's channel and sends them to the user's connection
func (u *User) ListenMessages() {
	for {
		msg := <-u.C
		u.conn.Write([]byte(msg + "\n"))
	}
}

// user goes online, adds the user to the server's online map and broadcasts a message
func (u *User) Online() {
	// user online,add user to online map
	u.server.mapLock.Lock()
	u.server.OnlineMap[u.Name] = u
	u.server.mapLock.Unlock()

	// broadcast message to other users
	u.server.Broadcast(u, "已上线")
}

// user goes offline, removes the user from the server's online map and broadcasts a message
func (u *User) Offline() {
	u.server.mapLock.Lock()
	delete(u.server.OnlineMap, u.Name)
	u.server.mapLock.Unlock()

	// broadcast message to other users
	u.server.Broadcast(u, "已下线")
}

// user handles a message
func (u *User) DoMessage(msg string) {
	if msg == "/who" { // list online users
		u.server.mapLock.Lock()
		for _, user := range u.server.OnlineMap {
			onlineMsg := "[" + user.Addr + "]" + user.Name + ":" + "在线..."
			u.C <- onlineMsg
		}
		u.server.mapLock.Unlock()
	} else if len(msg) > 7 && msg[:7] == "/rename" { // rename user: /rename newname
		newName := strings.Split(msg, " ")[1]
		// judge if the new name is already taken
		u.server.mapLock.Lock()
		_, ok := u.server.OnlineMap[newName]
		u.server.mapLock.Unlock()
		if ok {
			u.C <- "用户名已存在"
		} else {
			u.server.mapLock.Lock()
			delete(u.server.OnlineMap, u.Name)
			u.server.OnlineMap[newName] = u
			u.server.mapLock.Unlock()
			u.Name = newName
			u.C <- "用户名已修改为：" + newName
		}
	} else if len(msg) > 3 && msg[:3] == "/to" { // private message: /to username message
		parts := strings.SplitN(msg, " ", 3)
		if len(parts) < 3 {
			u.C <- "消息格式错误，请使用 /to username message 格式"
			return
		}

		targetName := strings.TrimSpace(parts[1])
		privateMsg := strings.TrimSpace(parts[2])

		if targetName == "" {
			u.C <- "消息格式错误，请使用 /to username message 格式"
			return
		}

		targetUser, ok := u.server.OnlineMap[targetName]
		if !ok {
			u.C <- "用户不存在"
			return
		}

		if privateMsg == "" {
			u.C <- "消息内容不能为空"
			return
		}
		targetUser.C <- "[" + u.Name + "]" + "对你说：" + privateMsg
	} else {
		u.server.Broadcast(u, msg)
	}

}
