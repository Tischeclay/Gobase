package InstantMessageSystem

type Server struct {
	Ip   string
	Port int
}

// 创建一个server的接口
func NewServer(ip string, port int) *Server {
	server := &Server{
		Ip:   ip,
		Port: port,
	}
	return server
}

func (s *Server) Start() {
	// socket listen

	// accept

	// do handler

	// close listen socket

}
