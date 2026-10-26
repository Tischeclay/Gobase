package main

import (
	"encoding/json"
	"log"
	"net/http"
	"time"
)

type Response struct {
	Message string `json:"message"`
}

func hello(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Query().Get("name")
	if name == "" {
		name = "世界"
	}

	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	if err := json.NewEncoder(w).Encode(Response{
		Message: "你好，" + name,
	}); err != nil {
		log.Printf("写入响应失败：%v", err)
	}
}

func main() {
	mux := http.NewServeMux()
	mux.HandleFunc("/hello", hello)

	server := &http.Server{
		Addr:              ":8080",
		Handler:           mux,
		ReadHeaderTimeout: 5 * time.Second,
	}

	log.Println("服务已启动：http://localhost:8080")
	log.Fatal(server.ListenAndServe())
}
