package main

import (
	"context"
	"log"
	"time"

	pb "grpc-demo/proto" // 替换为你的实际 module 名称

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	// 1. 建立与服务端的连接（这里选择明文/不安全传输模式）
	conn, err := grpc.Dial("localhost:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("连接服务端失败: %v", err)
	}
	defer conn.Close()

	// 2. 实例化客户端
	c := pb.NewGreeterClient(conn)

	// 3. 设置请求超时时间
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	// 4. 调用 RPC 方法
	r, err := c.SayHello(ctx, &pb.HelloRequest{Name: "World"})
	if err != nil {
		log.Fatalf("调用 RPC 失败: %v", err)
	}

	log.Printf("收到服务端响应: %s", r.GetMessage())
}
