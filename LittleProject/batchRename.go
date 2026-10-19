package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	dir := "./images" // 目标文件夹
	files, err := os.ReadDir(dir)
	if err != nil {
		fmt.Printf("读取目录失败: %v\n", err)
		return
	}

	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(file.Name(), ".png") {
			info, _ := file.Info()
			newPrefix := info.ModTime().Format("20060102_150405_")
			oldPath := filepath.Join(dir, file.Name())
			newPath := filepath.Join(dir, newPrefix+file.Name())

			os.Rename(oldPath, newPath)
			fmt.Printf("重命名: %s -> %s\n", file.Name(), newPrefix+file.Name())
		}
	}
}
