package main

import "net"

type User struct {
	Name string
	Addr string
	C    chan string
	conn net.Conn

	server *Server
}

func NewUser(conn net.Conn, server *Server) *User {
	userAddr := conn.RemoteAddr().String()

	user := &User{
		Name: userAddr,
		Addr: userAddr,
		C:    make(chan string),
		conn: conn,

		server: server,
	}
	go user.ListenMessage()
	return user
}

func (s *User) Online() {
	s.server.mapLock.Lock()
	s.server.OnlineMap[s.Name] = s
	s.server.mapLock.Unlock()
	s.server.Broadcast(s, "已上线")
}

func (s *User) Offline() {
	s.server.mapLock.Lock()
	delete(s.server.OnlineMap, s.Name)
	s.server.mapLock.Unlock()
	s.server.Broadcast(s, "下线")
}

func (s *User) DoMessage(msg string) {
	s.server.Broadcast(s, msg)
}

func (s *User) ListenMessage() {
	for {
		msg := <-s.C
		s.conn.Write([]byte(msg + "\n"))
	}
}
