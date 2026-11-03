package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// readMarkdown 读取 Markdown 文件，并逐行打印。
func readMarkdown(path string) error {
	if !strings.EqualFold(filepath.Ext(path), ".md") {
		return fmt.Errorf("请选择 .md 文件")
	}

	file, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("打开文件失败：%w", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)

	// Scanner 默认单行长度有限，这里将上限提高到 1 MiB。
	scanner.Buffer(make([]byte, 64*1024), 1024*1024)

	for scanner.Scan() {
		if _, err := fmt.Fprintln(os.Stdout, scanner.Text()); err != nil {
			return fmt.Errorf("打印文本失败：%w", err)
		}
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("读取文件失败：%w", err)
	}

	return nil
}

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintln(os.Stderr, "用法：go run main.go 文件路径.md")
		os.Exit(1)
	}

	if err := readMarkdown(os.Args[1]); err != nil {
		fmt.Fprintln(os.Stderr, "错误：", err)
		os.Exit(1)
	}
}
