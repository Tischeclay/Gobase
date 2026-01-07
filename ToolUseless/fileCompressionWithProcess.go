package main

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

type CompressProgress struct {
	TotalFiles  int
	Processed   int
	CurrentFile string
}

// 带进度显示的压缩函数
func CompressFolderWithProgress(sourceDir, destZip string, progress chan<- CompressProgress) error {
	defer close(progress)

	// 首先统计文件总数
	var totalFiles int
	filepath.Walk(sourceDir, func(path string, info os.FileInfo, err error) error {
		if !info.IsDir() {
			totalFiles++
		}
		return nil
	})

	zipFile, err := os.Create(destZip)
	if err != nil {
		return fmt.Errorf("创建压缩文件失败: %v", err)
	}
	defer zipFile.Close()

	zipWriter := zip.NewWriter(zipFile)
	defer zipWriter.Close()

	processed := 0

	err = filepath.Walk(sourceDir, func(filePath string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(sourceDir, filePath)
		if err != nil {
			return err
		}

		if info.IsDir() {
			relPath += "/"
			zipWriter.Create(relPath)
			return nil
		}

		progress <- CompressProgress{
			TotalFiles:  totalFiles,
			Processed:   processed,
			CurrentFile: relPath,
		}

		file, err := os.Open(filePath)
		if err != nil {
			return err
		}
		defer file.Close()

		header, err := zip.FileInfoHeader(info)
		if err != nil {
			return err
		}

		header.Name = relPath
		header.Method = zip.Deflate

		writer, err := zipWriter.CreateHeader(header)
		if err != nil {
			return err
		}

		_, err = io.Copy(writer, file)
		processed++

		progress <- CompressProgress{
			TotalFiles:  totalFiles,
			Processed:   processed,
			CurrentFile: relPath,
		}

		return err
	})

	return err
}

func main() {
	if len(os.Args) != 4 {
		fmt.Println("用法: program compress <源文件夹> <目标zip文件>")
		return
	}

	command := os.Args[1]
	source := os.Args[2]
	destination := os.Args[3]

	if command != "compress" {
		fmt.Println("未知命令")
		return
	}

	// 检查源文件夹是否存在
	if _, err := os.Stat(source); os.IsNotExist(err) {
		fmt.Printf("错误: 源文件夹 '%s' 不存在\n", source)
		return
	}

	if !strings.HasSuffix(destination, ".zip") {
		destination += ".zip"
	}

	progress := make(chan CompressProgress)

	go func() {
		for p := range progress {
			percent := float64(p.Processed) / float64(p.TotalFiles) * 100
			fmt.Printf("\r进度: %.1f%% (%d/%d) - 当前文件: %s", percent, p.Processed, p.TotalFiles, p.CurrentFile)
		}
		fmt.Println("\n压缩完成!")
	}()

	err := CompressFolderWithProgress(source, destination, progress)
	if err != nil {
		fmt.Printf("\n压缩失败: %v\n", err)
	}
}
