package server

import "testing"

func TestPrivateMessagePreservesSpaces(t *testing.T) {
	server := &Server{
		OnlineMap: make(map[string]*User),
		Message:   make(chan string),
	}

	sender := &User{Name: "alice", C: make(chan string, 1), server: server}
	target := &User{Name: "bob", C: make(chan string, 1), server: server}
	server.OnlineMap[target.Name] = target

	sender.DoMessage("/to bob hello world")

	got := <-target.C
	want := "[alice]对你说：hello world"
	if got != want {
		t.Fatalf("private message = %q, want %q", got, want)
	}
}
