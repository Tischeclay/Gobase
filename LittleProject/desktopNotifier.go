// notifier.go
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gen2brain/beeep"
	"github.com/gorilla/websocket"
)

type Notification struct {
	Title   string `json:"title"`
	Message string `json:"message"`
	Icon    string `json:"icon"`
	URL     string `json:"url"`
	Sound   bool   `json:"sound"`
}

type WebhookPayload struct {
	Event     string                 `json:"event"`
	Data      map[string]interface{} `json:"data"`
	Timestamp time.Time              `json:"timestamp"`
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

func main() {
	// 命令行参数
	port := flag.Int("port", 8080, "HTTP服务端口")
	webhookPath := flag.String("webhook", "/webhook", "Webhook路径")
	wsPath := flag.String("ws", "/ws", "WebSocket路径")
	flag.Parse()

	// 通知通道
	notifications := make(chan Notification, 100)

	// HTTP服务器
	http.HandleFunc(*webhookPath, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var payload WebhookPayload
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		// 转换为通知
		notif := Notification{
			Title:   fmt.Sprintf("事件: %s", payload.Event),
			Message: fmt.Sprintf("收到Webhook，数据: %v", payload.Data),
			Sound:   true,
		}

		notifications <- notif
		w.WriteHeader(http.StatusAccepted)
	})

	http.HandleFunc(*wsPath, func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Print("升级失败:", err)
			return
		}
		defer conn.Close()

		for {
			var notif Notification
			err := conn.ReadJSON(&notif)
			if err != nil {
				break
			}
			notifications <- notif
		}
	})

	// 健康检查
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// 启动服务器
	go func() {
		log.Printf("通知服务启动在 :%d", *port)
		log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", *port), nil))
	}()

	// 处理通知
	go func() {
		ticker := time.NewTicker(1 * time.Minute)
		defer ticker.Stop()

		for {
			select {
			case notif := <-notifications:
				// 显示通知
				err := beeep.Notify(notif.Title, notif.Message, notif.Icon)
				if err != nil {
					log.Printf("通知失败: %v", err)
				}

				// 播放声音
				if notif.Sound {
					beeep.Beep(beeep.DefaultFreq, beeep.DefaultDuration)
				}

				// 打开URL（如果有）
				if notif.URL != "" {
					beeep.Notify("正在打开链接", notif.URL, "")
				}

			case <-ticker.C:
				// 定期提醒
				beeep.Notify("提醒", "服务运行中", "")
			}
		}
	}()

	// 等待退出信号
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan
	log.Println("正在关闭...")
}
