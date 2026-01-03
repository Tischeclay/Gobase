package main

import (
	"context"
	"fmt"
	hello_grpc "github/piexlMax/gorm_class/grpc_class/pb"
	"google.golang.org/grpc"
	"net"
)

type server struct {
	hello_grpc.UnimplementedHelloGRPCServer
}

func (s *server) SayHi(ctx context.Context, req *hello_grpc.Req) (res *hello_grpc.Res, err error) {
	fmt.Println(req.GetMessage())
	return &hello_grpc.Res{Message: "我是从服务端返回的grpc的内容"}, nil
}

func main() {
	l, _ := net.Listen("tcp", ":8888")
	s := grpc.NewServer()
	hello_grpc.RegisterHelloGRPCServer(s, &server{})
	s.Serve(l)
}
