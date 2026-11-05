package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// analyzeLog 逐行读取日志，打印错误行并统计数量。
func analyzeLog(path string) error {
	file, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("打开日志失败：%w", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)

	// 将单行长度上限设为 1 MiB，避免较长日志触发默认限制。
	scanner.Buffer(make([]byte, 64*1024), 1024*1024)

	totalLines := 0
	errorLines := 0

	for scanner.Scan() {
		totalLines++
		line := scanner.Text()

		// 转成大写后匹配，因此也能识别 error、Error。
		if strings.Contains(strings.ToUpper(line), "ERROR") {
			errorLines++
			fmt.Printf("第 %d 行：%s\n", totalLines, line)
		}
	}

	// Scan 返回 false 也可能是读取失败，需要检查错误。
	if err := scanner.Err(); err != nil {
		return fmt.Errorf("读取日志失败：%w", err)
	}

	fmt.Printf("\n统计完成：共 %d 行，其中错误日志 %d 行\n",
		totalLines, errorLines)

	return nil
}

func main() {
	// os.Args[0] 是程序名，os.Args[1] 是用户传入的文件路径。
	if len(os.Args) != 2 {
		fmt.Fprintln(os.Stderr, "用法：go run main.go 日志文件")
		os.Exit(1)
	}

	if err := analyzeLog(os.Args[1]); err != nil {
		fmt.Fprintln(os.Stderr, "错误：", err)
		os.Exit(1)
	}
}
