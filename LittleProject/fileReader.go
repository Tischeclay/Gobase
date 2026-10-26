package main

import (
	"fmt"
	"io"
	"os"
)

func copyFile(src, dst string) error {
	input, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("打开源文件：%w", err)
	}
	defer input.Close()

	info, err := input.Stat()
	if err != nil {
		return err
	}
	if !info.Mode().IsRegular() {
		return fmt.Errorf("源路径必须是普通文件")
	}

	output, err := os.OpenFile(
		dst,
		os.O_WRONLY|os.O_CREATE|os.O_EXCL,
		info.Mode().Perm(),
	)
	if err != nil {
		return fmt.Errorf("创建目标文件：%w", err)
	}

	_, copyErr := io.Copy(output, input)
	closeErr := output.Close()

	if copyErr != nil {
		os.Remove(dst)
		return fmt.Errorf("复制文件：%w", copyErr)
	}
	return closeErr
}

func main() {
	if len(os.Args) != 3 {
		fmt.Println("用法：go run copy.go 源文件 目标文件")
		os.Exit(1)
	}

	if err := copyFile(os.Args[1], os.Args[2]); err != nil {
		fmt.Fprintln(os.Stderr, "复制失败：", err)
		os.Exit(1)
	}

	fmt.Println("复制成功")
}
