package main

import (
	"archive/zip"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

func compressFile(sourcePath, outputPath string) error {
	// 创建输出文件
	outFile, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("创建压缩文件失败: %v", err)
	}
	defer outFile.Close()

	// 创建 ZIP writer
	zipWriter := zip.NewWriter(outFile)
	defer zipWriter.Close()

	// 打开源文件
	sourceFile, err := os.Open(sourcePath)
	if err != nil {
		return fmt.Errorf("打开源文件失败: %v", err)
	}
	defer sourceFile.Close()

	// 获取文件信息
	fileInfo, err := sourceFile.Stat()
	if err != nil {
		return fmt.Errorf("获取文件信息失败: %v", err)
	}

	// 创建 ZIP 文件头
	header, err := zip.FileInfoHeader(fileInfo)
	if err != nil {
		return fmt.Errorf("创建 ZIP 文件头失败: %v", err)
	}

	// 设置压缩方法
	header.Method = zip.Deflate
	header.Name = filepath.Base(sourcePath)

	// 创建 ZIP 文件写入器
	writer, err := zipWriter.CreateHeader(header)
	if err != nil {
		return fmt.Errorf("创建 ZIP 写入器失败: %v", err)
	}

	// 复制文件内容到 ZIP
	_, err = io.Copy(writer, sourceFile)
	if err != nil {
		return fmt.Errorf("写入 ZIP 文件失败: %v", err)
	}

	return nil
}

func processDirectory(targetPath, outputDir string, recursive bool) error {
	// 确保输出目录存在
	if outputDir != "" {
		err := os.MkdirAll(outputDir, 0755)
		if err != nil {
			return fmt.Errorf("创建输出目录失败: %v", err)
		}
	}

	// 统计信息
	var successCount, failCount int
	var errors []string

	// 遍历目录
	err := filepath.Walk(targetPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// 跳过目录（除非是递归模式且是根目录）
		if info.IsDir() {
			if !recursive || path == targetPath {
				return nil
			}
			return filepath.SkipDir
		}

		// 跳过已经是 ZIP 文件的情况（可选）
		if strings.HasSuffix(strings.ToLower(path), ".zip") {
			fmt.Printf("跳过 ZIP 文件: %s\n", path)
			return nil
		}

		// 生成输出文件名
		baseName := filepath.Base(path)
		ext := filepath.Ext(baseName)
		zipName := strings.TrimSuffix(baseName, ext) + ".zip"

		var outputPath string
		if outputDir != "" {
			outputPath = filepath.Join(outputDir, zipName)
		} else {
			// 输出到源文件同目录
			outputPath = filepath.Join(filepath.Dir(path), zipName)
		}

		// 检查输出文件是否已存在
		if _, err := os.Stat(outputPath); err == nil {
			fmt.Printf("跳过已存在的文件: %s\n", outputPath)
			return nil
		}

		// 压缩文件
		fmt.Printf("正在压缩: %s -> %s\n", path, outputPath)
		err = compressFile(path, outputPath)
		if err != nil {
			failCount++
			errorMsg := fmt.Sprintf("%s: %v", path, err)
			errors = append(errors, errorMsg)
			fmt.Printf("  ❌ 失败: %v\n", err)
			return nil // 继续处理下一个文件
		}

		successCount++
		fmt.Printf("  ✅ 成功\n")
		return nil
	})

	if err != nil {
		return fmt.Errorf("遍历目录失败: %v", err)
	}

	// 打印统计信息
	fmt.Printf("\n=== 处理完成 ===\n")
	fmt.Printf("成功: %d 个文件\n", successCount)
	fmt.Printf("失败: %d 个文件\n", failCount)

	if len(errors) > 0 {
		fmt.Printf("\n错误详情:\n")
		for _, errMsg := range errors {
			fmt.Printf("  - %s\n", errMsg)
		}
	}

	return nil
}

func main() {
	var (
		targetPath = flag.String("path", "", "目标路径（文件或目录）")
		outputDir  = flag.String("output", "", "输出目录（为空则输出到源文件同目录）")
		recursive  = flag.Bool("recursive", false, "递归处理子目录")
		help       = flag.Bool("help", false, "显示帮助信息")
	)

	flag.Parse()

	if *help || *targetPath == "" {
		fmt.Println("文件压缩批处理程序")
		fmt.Println("\n用法:")
		fmt.Println("  fileCompressor -path <目标路径> [选项]")
		fmt.Println("\n选项:")
		flag.PrintDefaults()
		fmt.Println("\n示例:")
		fmt.Println("  fileCompressor -path ./documents")
		fmt.Println("  fileCompressor -path ./documents -output ./compressed")
		fmt.Println("  fileCompressor -path ./documents -recursive")
		fmt.Println("  fileCompressor -path ./documents -output ./compressed -recursive")
		os.Exit(0)
	}

	// 检查目标路径是否存在
	info, err := os.Stat(*targetPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "错误: 目标路径不存在: %v\n", err)
		os.Exit(1)
	}

	// 如果是单个文件，直接压缩
	if !info.IsDir() {
		baseName := filepath.Base(*targetPath)
		ext := filepath.Ext(baseName)
		zipName := strings.TrimSuffix(baseName, ext) + ".zip"

		var outputPath string
		if *outputDir != "" {
			os.MkdirAll(*outputDir, 0755)
			outputPath = filepath.Join(*outputDir, zipName)
		} else {
			outputPath = filepath.Join(filepath.Dir(*targetPath), zipName)
		}

		fmt.Printf("正在压缩: %s -> %s\n", *targetPath, outputPath)
		err = compressFile(*targetPath, outputPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "压缩失败: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("✅ 压缩完成")
		os.Exit(0)
	}

	// 处理目录
	err = processDirectory(*targetPath, *outputDir, *recursive)
	if err != nil {
		fmt.Fprintf(os.Stderr, "处理失败: %v\n", err)
		os.Exit(1)
	}
}
