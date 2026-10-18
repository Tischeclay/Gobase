package main

import (
	"fmt"
	"net"
	"sync"
	"time"
)

func checkPort(ip string, port int, wg *sync.WaitGroup) {
	defer wg.Done()
	address := fmt.Sprintf("%s:%d", ip, port)
	conn, err := net.DialTimeout("tcp", address, 2*time.Second)
	if err != nil {
		return // 端口关闭或不可达
	}
	conn.Close()
	fmt.Printf("[+] 端口开放: %s\n", address)
}

func main() {
	target := "127.0.0.1"
	ports := []int{22, 80, 443, 3306, 6379, 8080}
	var wg sync.WaitGroup

	fmt.Printf("开始扫描目标 %s...\n", target)
	for _, p := range ports {
		wg.Add(1)
		go checkPort(target, p, &wg)
	}
	wg.Wait()
	fmt.Println("扫描完成。")
}
