//package main
//
//import (
//	"context"
//	"log"
//	"time"
//
//	pb "grpc-demo/proto" // 替换为你的实际 module 名称
//
//	"google.golang.org/grpc"
//	"google.golang.org/grpc/credentials/insecure"
//)
//
//func main() {
//	// 1. 建立与服务端的连接（这里选择明文/不安全传输模式）
//	conn, err := grpc.Dial("localhost:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
//	if err != nil {
//		log.Fatalf("连接服务端失败: %v", err)
//	}
//	defer conn.Close()
//
//	// 2. 实例化客户端
//	c := pb.NewGreeterClient(conn)
//
//	// 3. 设置请求超时时间
//	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
//	defer cancel()
//
//	// 4. 调用 RPC 方法
//	r, err := c.SayHello(ctx, &pb.HelloRequest{Name: "World"})
//	if err != nil {
//		log.Fatalf("调用 RPC 失败: %v", err)
//	}
//
//	log.Printf("收到服务端响应: %s", r.GetMessage())
//}

package main

import (
	"context"
	"io"
	"log"
	"time"

	pb "grpc-demo/proto"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)

func main() {
	// 1. 连接服务端
	conn, err := grpc.Dial("localhost:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("连接服务端失败: %v", err)
	}
	defer conn.Close()

	client := pb.NewChatServiceClient(conn)

	// 2. 构造带 Token 的 Metadata 上下文
	md := metadata.Pairs("authorization", "Bearer secret-token-123")
	ctx := metadata.NewOutgoingContext(context.Background(), md)

	// 3. 打开双向流
	stream, err := client.BiStreamChat(ctx)
	if err != nil {
		log.Fatalf("创建流失败: %v", err)
	}

	waitc := make(chan struct{})

	// 4. 开启独立协程持续接收服务端推过来的消息
	go func() {
		for {
			in, err := stream.Recv()
			if err == io.EOF {
				// 服务端关闭了流
				close(waitc)
				return
			}
			if err != nil {
				log.Fatalf("接收服务端消息失败: %v", err)
			}
			log.Printf("[收到推流] %s (%s): %s\n",
				in.GetUser(),
				time.Unix(in.GetTimestamp(), 0).Format("15:04:05"),
				in.GetMessage(),
			)
		}
	}()

	// 5. 主线程模拟循环发送消息
	messages := []string{"你好！", "请问今天天气怎么样？", "好的，谢谢你！"}
	for _, msg := range messages {
		req := &pb.ChatMessage{
			User:      "Alice",
			Message:   msg,
			Timestamp: time.Now().Unix(),
		}
		if err := stream.Send(req); err != nil {
			log.Fatalf("发送消息失败: %v", err)
		}
		time.Sleep(1 * time.Second) // 模拟间隔
	}

	// 6. 关闭发送流，告知服务端发送完成
	stream.CloseSend()
	<-waitc // 等待接收协程退出
	log.Println("双向流会话结束")
}
