// chat_server.go
package main

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

type Client struct {
	ID       string
	Username string
	RoomID   string
	Conn     *websocket.Conn
	Send     chan Message
}

type Message struct {
	ID        string    `json:"id"`
	Type      string    `json:"type"` // "chat", "private", "system", "join", "leave"
	Content   string    `json:"content"`
	Sender    string    `json:"sender"`
	Receiver  string    `json:"receiver,omitempty"`
	RoomID    string    `json:"room_id"`
	Timestamp time.Time `json:"timestamp"`
}

type Room struct {
	ID      string
	Name    string
	Clients map[string]*Client
	History []Message
	mu      sync.RWMutex
}

type ChatServer struct {
	clients    map[string]*Client
	rooms      map[string]*Room
	register   chan *Client
	unregister chan *Client
	broadcast  chan Message
	mu         sync.RWMutex
	upgrader   websocket.Upgrader
}

func NewChatServer() *ChatServer {
	return &ChatServer{
		clients:    make(map[string]*Client),
		rooms:      make(map[string]*Room),
		register:   make(chan *Client),
		unregister: make(chan *Client),
		broadcast:  make(chan Message),
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool { return true },
		},
	}
}

func (s *ChatServer) Run() {
	for {
		select {
		case client := <-s.register:
			s.mu.Lock()
			s.clients[client.ID] = client

			// 加入房间
			if room, ok := s.rooms[client.RoomID]; ok {
				room.mu.Lock()
				room.Clients[client.ID] = client
				room.mu.Unlock()

				// 发送历史消息
				room.mu.RLock()
				for _, msg := range room.History {
					if len(room.History) > 50 {
						break
					}
					client.Send <- msg
				}
				room.mu.RUnlock()
			}
			s.mu.Unlock()

			// 广播加入消息
			s.broadcast <- Message{
				Type:      "system",
				Content:   client.Username + " 加入了聊天室",
				RoomID:    client.RoomID,
				Timestamp: time.Now(),
			}

		case client := <-s.unregister:
			s.mu.Lock()
			if _, ok := s.clients[client.ID]; ok {
				delete(s.clients, client.ID)

				if room, ok := s.rooms[client.RoomID]; ok {
					room.mu.Lock()
					delete(room.Clients, client.ID)
					room.mu.Unlock()
				}

				close(client.Send)

				s.broadcast <- Message{
					Type:      "system",
					Content:   client.Username + " 离开了聊天室",
					RoomID:    client.RoomID,
					Timestamp: time.Now(),
				}
			}
			s.mu.Unlock()

		case message := <-s.broadcast:
			s.mu.RLock()
			room, ok := s.rooms[message.RoomID]
			s.mu.RUnlock()

			if !ok {
				continue
			}

			// 保存到历史
			room.mu.Lock()
			room.History = append([]Message{message}, room.History...)
			if len(room.History) > 100 {
				room.History = room.History[:100]
			}
			room.mu.Unlock()

			// 广播给房间内所有客户端
			room.mu.RLock()
			for _, client := range room.Clients {
				if message.Type == "private" && client.ID != message.Receiver {
					continue
				}

				select {
				case client.Send <- message:
				default:
					close(client.Send)
					delete(room.Clients, client.ID)
				}
			}
			room.mu.RUnlock()
		}
	}
}

func (s *ChatServer) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("WebSocket upgrade failed:", err)
		return
	}

	username := r.URL.Query().Get("username")
	roomID := r.URL.Query().Get("room")

	if username == "" {
		username = "Anonymous"
	}
	if roomID == "" {
		roomID = "general"
	}

	// 确保房间存在
	s.mu.Lock()
	if _, ok := s.rooms[roomID]; !ok {
		s.rooms[roomID] = &Room{
			ID:      roomID,
			Name:    roomID,
			Clients: make(map[string]*Client),
			History: make([]Message, 0),
		}
	}
	s.mu.Unlock()

	client := &Client{
		ID:       uuid.New().String(),
		Username: username,
		RoomID:   roomID,
		Conn:     conn,
		Send:     make(chan Message, 256),
	}

	s.register <- client

	// 处理发送消息
	go func() {
		defer func() {
			s.unregister <- client
			conn.Close()
		}()

		for {
			var msg Message
			err := conn.ReadJSON(&msg)
			if err != nil {
				break
			}

			msg.ID = uuid.New().String()
			msg.Sender = client.Username
			msg.RoomID = client.RoomID
			msg.Timestamp = time.Now()

			if msg.Type == "chat" || msg.Type == "private" {
				s.broadcast <- msg
			}
		}
	}()

	// 处理接收消息
	go func() {
		for msg := range client.Send {
			err := conn.WriteJSON(msg)
			if err != nil {
				break
			}
		}
	}()
}

func (s *ChatServer) HandleRooms(w http.ResponseWriter, r *http.Request) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	roomList := make([]map[string]interface{}, 0)
	for id, room := range s.rooms {
		room.mu.RLock()
		roomList = append(roomList, map[string]interface{}{
			"id":         id,
			"name":       room.Name,
			"user_count": len(room.Clients),
		})
		room.mu.RUnlock()
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(roomList)
}

func (s *ChatServer) HandleRoomHistory(w http.ResponseWriter, r *http.Request) {
	roomID := r.URL.Query().Get("room")

	s.mu.RLock()
	room, ok := s.rooms[roomID]
	s.mu.RUnlock()

	if !ok {
		http.Error(w, "Room not found", http.StatusNotFound)
		return
	}

	room.mu.RLock()
	history := make([]Message, len(room.History))
	copy(history, room.History)
	room.mu.RUnlock()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(history)
}

func (s *ChatServer) HandleRoomUsers(w http.ResponseWriter, r *http.Request) {
	roomID := r.URL.Query().Get("room")

	s.mu.RLock()
	room, ok := s.rooms[roomID]
	s.mu.RUnlock()

	if !ok {
		http.Error(w, "Room not found", http.StatusNotFound)
		return
	}

	room.mu.RLock()
	users := make([]string, 0, len(room.Clients))
	for _, client := range room.Clients {
		users = append(users, client.Username)
	}
	room.mu.RUnlock()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(users)
}

func main() {
	server := NewChatServer()
	go server.Run()

	// 静态文件服务
	http.Handle("/", http.FileServer(http.Dir("static")))

	// WebSocket路由
	http.HandleFunc("/ws", server.HandleWebSocket)
	http.HandleFunc("/api/rooms", server.HandleRooms)
	http.HandleFunc("/api/history", server.HandleRoomHistory)
	http.HandleFunc("/api/users", server.HandleRoomUsers)

	log.Println("Chat server starting on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
