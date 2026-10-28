package main

import (
	"crypto/sha256"
	"fmt"
	"io"
	"os"
)

// 流式读取文件，避免把整个文件加载到内存。
func checksum(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", fmt.Errorf("打开文件：%w", err)
	}
	defer file.Close()

	info, err := file.Stat()
	if err != nil {
		return "", fmt.Errorf("读取文件信息：%w", err)
	}
	if !info.Mode().IsRegular() {
		return "", fmt.Errorf("%s 不是普通文件", path)
	}

	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", fmt.Errorf("读取文件：%w", err)
	}

	return fmt.Sprintf("%x", hash.Sum(nil)), nil
}

func main() {
	if len(os.Args) < 2 || len(os.Args) > 3 {
		fmt.Fprintln(os.Stderr, "用法：go run checksum.go 文件 [另一个文件]")
		os.Exit(1)
	}

	first, err := checksum(os.Args[1])
	if err != nil {
		fmt.Fprintln(os.Stderr, "计算失败：", err)
		os.Exit(1)
	}
	fmt.Printf("%s  %s\n", first, os.Args[1])

	if len(os.Args) == 2 {
		return
	}

	second, err := checksum(os.Args[2])
	if err != nil {
		fmt.Fprintln(os.Stderr, "计算失败：", err)
		os.Exit(1)
	}
	fmt.Printf("%s  %s\n", second, os.Args[2])

	if first != second {
		fmt.Println("两个文件的内容不同")
		os.Exit(1)
	}

	fmt.Println("两个文件的 SHA-256 相同，可用于判断内容一致")
}
