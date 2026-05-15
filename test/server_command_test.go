package test

import (
	"io"
	"net"
	"reflect"
	"sort"
	"strings"
	"testing"
	"time"

	imserver "github.com/Yuyuzheng2499/mini-go-im/internal/server"
)

type testClient struct {
	user *imserver.User
	conn *recordingConn
}

type recordingConn struct {
	messages chan string
}

func newRecordingConn() *recordingConn {
	return &recordingConn{messages: make(chan string, 100)}
}

func (c *recordingConn) Read([]byte) (int, error) {
	return 0, io.EOF
}

func (c *recordingConn) Write(p []byte) (int, error) {
	c.messages <- strings.TrimSpace(string(p))
	return len(p), nil
}

func (c *recordingConn) Close() error {
	return nil
}

func (c *recordingConn) LocalAddr() net.Addr {
	return testAddr("local")
}

func (c *recordingConn) RemoteAddr() net.Addr {
	return testAddr("remote")
}

func (c *recordingConn) SetDeadline(time.Time) error {
	return nil
}

func (c *recordingConn) SetReadDeadline(time.Time) error {
	return nil
}

func (c *recordingConn) SetWriteDeadline(time.Time) error {
	return nil
}

type testAddr string

func (a testAddr) Network() string {
	return "test"
}

func (a testAddr) String() string {
	return string(a)
}

func newTestClient(t *testing.T, srv *imserver.Server, name string) *testClient {
	t.Helper()

	conn := newRecordingConn()
	user := imserver.NewUser(conn, srv)
	user.Name = name
	user.Addr = name + "-addr"
	srv.OnlineMap[user.Name] = user

	return &testClient{user: user, conn: conn}
}

func readLine(t *testing.T, client *testClient) string {
	t.Helper()

	select {
	case msg := <-client.conn.messages:
		return msg
	case <-time.After(500 * time.Millisecond):
		t.Fatalf("read response for %s timed out", client.user.Name)
	}

	return ""
}

func assertNoLine(t *testing.T, client *testClient) {
	t.Helper()

	select {
	case msg := <-client.conn.messages:
		t.Fatalf("unexpected response for %s: %q", client.user.Name, msg)
	case <-time.After(50 * time.Millisecond):
	}
}

func TestWhoCommandReturnsOnlineUsersOnlyToRequester(t *testing.T) {
	srv := imserver.NewServer("127.0.0.1", 0)
	alice := newTestClient(t, srv, "alice")
	bob := newTestClient(t, srv, "bob")

	alice.user.DoMessage("/who")

	got := []string{readLine(t, alice), readLine(t, alice)}
	sort.Strings(got)

	want := []string{
		"[alice-addr]alice:在线...",
		"[bob-addr]bob:在线...",
	}
	sort.Strings(want)

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("/who response = %#v, want %#v", got, want)
	}

	assertNoLine(t, bob)
}

func TestRenameCommandUpdatesNameAndOnlineMap(t *testing.T) {
	srv := imserver.NewServer("127.0.0.1", 0)
	alice := newTestClient(t, srv, "alice")

	alice.user.DoMessage("/rename 张三")

	if got, want := readLine(t, alice), "用户名已修改为：张三"; got != want {
		t.Fatalf("rename response = %q, want %q", got, want)
	}

	if alice.user.Name != "张三" {
		t.Fatalf("user name = %q, want %q", alice.user.Name, "张三")
	}
	if _, ok := srv.OnlineMap["alice"]; ok {
		t.Fatal("old user name still exists in online map")
	}
	if got := srv.OnlineMap["张三"]; got != alice.user {
		t.Fatalf("online map entry for new name = %p, want %p", got, alice.user)
	}
}

func TestRenameCommandRejectsDuplicateName(t *testing.T) {
	srv := imserver.NewServer("127.0.0.1", 0)
	alice := newTestClient(t, srv, "alice")
	bob := newTestClient(t, srv, "bob")

	alice.user.DoMessage("/rename bob")

	if got, want := readLine(t, alice), "用户名已存在"; got != want {
		t.Fatalf("duplicate rename response = %q, want %q", got, want)
	}
	if alice.user.Name != "alice" {
		t.Fatalf("user name after duplicate rename = %q, want %q", alice.user.Name, "alice")
	}
	if srv.OnlineMap["alice"] != alice.user {
		t.Fatal("original user was removed from online map")
	}
	if srv.OnlineMap["bob"] != bob.user {
		t.Fatal("duplicate target user was changed in online map")
	}
}

func TestPrivateMessagePreservesSpacesAndOnlyTargetsRecipient(t *testing.T) {
	srv := imserver.NewServer("127.0.0.1", 0)
	alice := newTestClient(t, srv, "alice")
	bob := newTestClient(t, srv, "bob")

	alice.user.DoMessage("/to bob hello world")

	if got, want := readLine(t, bob), "[alice]对你说：hello world"; got != want {
		t.Fatalf("private message = %q, want %q", got, want)
	}
	assertNoLine(t, alice)
}

func TestPrivateMessageValidationResponses(t *testing.T) {
	tests := []struct {
		name      string
		msg       string
		want      string
		createBob bool
	}{
		{
			name: "missing message part",
			msg:  "/to bob",
			want: "消息格式错误，请使用 /to username message 格式",
		},
		{
			name: "missing target name",
			msg:  "/to   hello",
			want: "消息格式错误，请使用 /to username message 格式",
		},
		{
			name: "unknown target user",
			msg:  "/to bob hello",
			want: "用户不存在",
		},
		{
			name:      "empty message content",
			msg:       "/to bob    ",
			want:      "消息内容不能为空",
			createBob: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv := imserver.NewServer("127.0.0.1", 0)
			alice := newTestClient(t, srv, "alice")
			if tt.createBob {
				newTestClient(t, srv, "bob")
			}

			alice.user.DoMessage(tt.msg)

			if got := readLine(t, alice); got != tt.want {
				t.Fatalf("validation response = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestBroadcastDeliversRegularMessagesToAllOnlineUsers(t *testing.T) {
	srv := imserver.NewServer("127.0.0.1", 0)
	go srv.ListenMessages()

	alice := newTestClient(t, srv, "alice")
	bob := newTestClient(t, srv, "bob")

	alice.user.DoMessage("hello")

	want := "[alice-addr]:alice:hello"
	if got := readLine(t, alice); got != want {
		t.Fatalf("sender broadcast = %q, want %q", got, want)
	}
	if got := readLine(t, bob); got != want {
		t.Fatalf("recipient broadcast = %q, want %q", got, want)
	}
}

func TestListenMessagesWritesChannelMessagesToConnection(t *testing.T) {
	srv := imserver.NewServer("127.0.0.1", 0)
	alice := newTestClient(t, srv, "alice")
	done := make(chan struct{})

	go func() {
		alice.user.C <- "系统消息"
		close(done)
	}()

	if got, want := readLine(t, alice), "系统消息"; got != want {
		t.Fatalf("channel message = %q, want %q", got, want)
	}

	select {
	case <-done:
	case <-time.After(500 * time.Millisecond):
		t.Fatal("sending channel message did not finish")
	}
}
