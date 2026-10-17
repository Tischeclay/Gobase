package main

import (
	"context"
	"fmt"
	"log"
	"net"

	pb "grpc-demo/proto" // 替换为你的实际 module 名称

	"google.golang.org/grpc"
)

// server 用于实现 proto 中定义的 GreeterServer 接口
type server struct {
	pb.UnimplementedGreeterServer
}

// SayHello 实现 GreeterServer 的 SayHello 方法
func (s *server) SayHello(ctx context.Context, in *pb.HelloRequest) (*pb.HelloReply, error) {
	log.Printf("收到客户端请求: %v", in.GetName())
	return &pb.HelloReply{Message: "Hello " + in.GetName()}, nil
}

func main() {
	// 1. 监听端口
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("监听端口失败: %v", err)
	}

	// 2. 创建 gRPC 服务器实例
	grpcServer := grpc.NewServer()

	// 3. 注册服务
	pb.RegisterGreeterServer(grpcServer, &server{})

	fmt.Println("gRPC 服务端启动，监听端口 :50051...")

	// 4. 启动服务
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("启动服务失败: %v", err)
	}
}
