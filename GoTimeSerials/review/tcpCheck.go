package main

import (
	"flag"
	"fmt"
	"net"
	"os"
	"strconv"
	"time"
)

// checkPort 尝试建立 TCP 连接，返回连接耗时。
func checkPort(host string, port int, timeout time.Duration) (time.Duration, error) {
	// JoinHostPort 能正确拼接普通域名、IPv4 和 IPv6 地址。
	address := net.JoinHostPort(host, strconv.Itoa(port))

	start := time.Now()

	// 在规定时间内尝试连接，避免一直等待。
	conn, err := net.DialTimeout("tcp", address, timeout)
	elapsed := time.Since(start)

	if err != nil {
		return elapsed, fmt.Errorf("连接 %s 失败：%w", address, err)
	}

	// 连接成功后及时关闭，释放资源。
	conn.Close()

	return elapsed, nil
}

func run() int {
	// 定义命令行参数。
	host := flag.String("host", "127.0.0.1", "目标主机，不包含协议前缀")
	port := flag.Int("port", 8080, "目标端口")
	timeout := flag.Duration("timeout", 3*time.Second, "连接超时时间")
	flag.Parse()

	// 检查参数是否合法。
	if *host == "" {
		fmt.Fprintln(os.Stderr, "主机不能为空")
		return 1
	}
	if *port < 1 || *port > 65535 {
		fmt.Fprintln(os.Stderr, "端口必须在 1～65535 之间")
		return 1
	}
	if *timeout <= 0 {
		fmt.Fprintln(os.Stderr, "超时时间必须大于 0")
		return 1
	}

	elapsed, err := checkPort(*host, *port, *timeout)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 1
	}

	fmt.Printf(
		"连接成功：%s，耗时：%s\n",
		net.JoinHostPort(*host, strconv.Itoa(*port)),
		elapsed.Round(time.Microsecond),
	)

	return 0
}

func main() {
	os.Exit(run())
}
