package main

import (
	"log"
	"net/http"
	"os"
)

func main() {
	port := "8080"
	dir := "."
	if len(os.Args) > 1 {
		dir = os.Args[1]
	}
	log.Printf("服务已启动，正在提供 [%s] 目录的访问：http://localhost:%s\n", dir, port)
	log.Fatal(http.ListenAndServe(":"+port, http.FileServer(http.Dir(dir))))
}
