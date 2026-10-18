//package main
//
//import (
//	"context"
//	"fmt"
//	"log"
//	"net"
//
//	pb "grpc-demo/proto" // 替换为你的实际 module 名称
//
//	"google.golang.org/grpc"
//)
//
//// server 用于实现 proto 中定义的 GreeterServer 接口
//type server struct {
//	pb.UnimplementedGreeterServer
//}
//
//// SayHello 实现 GreeterServer 的 SayHello 方法
//func (s *server) SayHello(ctx context.Context, in *pb.HelloRequest) (*pb.HelloReply, error) {
//	log.Printf("收到客户端请求: %v", in.GetName())
//	return &pb.HelloReply{Message: "Hello " + in.GetName()}, nil
//}
//
//func main() {
//	// 1. 监听端口
//	lis, err := net.Listen("tcp", ":50051")
//	if err != nil {
//		log.Fatalf("监听端口失败: %v", err)
//	}
//
//	// 2. 创建 gRPC 服务器实例
//	grpcServer := grpc.NewServer()
//
//	// 3. 注册服务
//	pb.RegisterGreeterServer(grpcServer, &server{})
//
//	fmt.Println("gRPC 服务端启动，监听端口 :50051...")
//
//	// 4. 启动服务
//	if err := grpcServer.Serve(lis); err != nil {
//		log.Fatalf("启动服务失败: %v", err)
//	}
//}

package main

import (
	"fmt"
	"io"
	"log"
	"net"
	"time"

	pb "grpc-demo/proto"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

type chatServer struct {
	pb.UnimplementedChatServiceServer
}

// BiStreamChat 实现双向流服务
func (s *chatServer) BiStreamChat(stream pb.ChatService_BiStreamChatServer) error {
	// 1. 从流上下文提取 Metadata (Header)
	md, ok := metadata.FromIncomingContext(stream.Context())
	if !ok {
		return status.Error(codes.Unauthenticated, "缺失元数据")
	}

	tokens := md.Get("authorization")
	if len(tokens) == 0 || tokens[0] != "Bearer secret-token-123" {
		return status.Error(codes.Unauthenticated, "Token 验证失败，拒绝连接")
	}
	log.Println("[服务端] 鉴权成功，建立长连接交互")

	// 2. 接收客户端发来的消息并实时响应
	for {
		req, err := stream.Recv()
		if err == io.EOF {
			log.println("[服务端] 客户端已关闭发送流")
			return nil
		}
		if err != nil {
			log.Printf("[服务端] 接收流出错: %v\n", err)
			return err
		}

		log.Printf("[收到消息] %s: %s\n", req.GetUser(), req.GetMessage())

		// 回复客户端
		resp := &pb.ChatMessage{
			User:      "系统机器人",
			Message:   fmt.Sprintf("已收到你的消息: '%s'", req.GetMessage()),
			Timestamp: time.Now().Unix(),
		}

		if err := stream.Send(resp); err != nil {
			log.Printf("[服务端] 发送流出错: %v\n", err)
			return err
		}
	}
}

// StreamServerInterceptor 流式服务端的日志中间件
func StreamServerInterceptor(
	srv interface{},
	ss grpc.ServerStream,
	info *grpc.StreamServerMethodInfo,
	handler grpc.StreamHandler,
) error {
	log.Printf("[Interceptor] 开始建立连接方法: %s\n", info.FullMethod)
	err := handler(srv, ss)
	log.Printf("[Interceptor] 连接断开/完成: %s, 状态: %v\n", info.FullMethod, err)
	return err
}

func main() {
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("监听端口失败: %v", err)
	}

	// 注册带拦截器的 gRPC 服务
	grpcServer := grpc.NewServer(
		grpc.StreamInterceptor(StreamServerInterceptor),
	)

	pb.RegisterChatServiceServer(grpcServer, &chatServer{})

	fmt.Println("gRPC 双向流服务端已启动，监听 :50051...")
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("服务运行失败: %v", err)
	}
}
