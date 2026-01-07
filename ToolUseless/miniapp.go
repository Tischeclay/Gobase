package main

import (
	"fmt"      // 用于格式化IO
	"log"      // 用于日志记录
	"net/http" //用于HTTP处理
)

func helloHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "Hello World!")
}
func main() {
	http.HandleFunc("/", helloHandler) //注册处理函数log.Println("Listening on :8080")//日志输出err := http.ListenAndServe(":8080"，nil)//启动 HTTP 服务器if err != nil {
	log.Println("listening on : 8080") //如果发生错误，则记录并退出
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal(err)
	}
}
