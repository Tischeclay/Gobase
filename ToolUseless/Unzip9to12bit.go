package main

import (
	"bytes"
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"runtime"
	"sync"
	"sync/atomic"
	"time"

	"github.com/alexmullins/zip"
)

var (
	zipFile    = "AI代码.zip"
	minLength  = flag.Int("min", 9, "最小密码长度")
	maxLength  = flag.Int("max", 12, "最大密码长度")
	numWorkers = flag.Int("w", 0, "并发worker数量 (0=自动设置为CPU核心数)")
	charset    = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ!@#$%^&*()_+-=[]{}|;:,.<>?"
)

func main() {
	flag.Parse()

	if *minLength < 1 || *maxLength < *minLength {
		fmt.Fprintf(os.Stderr, "错误: 密码长度范围无效\n")
		os.Exit(1)
	}

	if *numWorkers <= 0 {
		*numWorkers = runtime.NumCPU()
	}

	fmt.Printf("开始破解ZIP文件: %s\n", zipFile)
	fmt.Printf("密码长度范围: %d-%d 位\n", *minLength, *maxLength)
	fmt.Printf("字符集大小: %d\n", len(charset))
	fmt.Printf("并发Worker数量: %d\n", *numWorkers)
	fmt.Println("正在破解中...")

	startTime := time.Now()
	password := crackZip(zipFile, *minLength, *maxLength, *numWorkers)
	elapsed := time.Since(startTime)

	if password != "" {
		fmt.Printf("\n✓ 破解成功!\n")
		fmt.Printf("密码: %s\n", password)
		fmt.Printf("耗时: %v\n", elapsed)
	} else {
		fmt.Printf("\n✗ 未找到密码 (尝试了所有可能的组合)\n")
		fmt.Printf("耗时: %v\n", elapsed)
		os.Exit(1)
	}
}

func crackZip(zipPath string, minLen, maxLen, workers int) string {
	// 读取ZIP文件到内存
	zipData, err := os.ReadFile(zipPath)
	if err != nil {
		log.Fatalf("无法读取ZIP文件: %v", err)
	}

	// 创建context用于取消操作
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 创建任务通道和结果通道
	taskChan := make(chan string, workers*2)
	resultChan := make(chan string, 1)
	var wg sync.WaitGroup
	var found int32 // 原子变量，标记是否已找到密码
	var count int64 // 已尝试的密码数量

	// 启动worker goroutines
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go worker(ctx, zipData, taskChan, resultChan, &wg, &found, &count)
	}

	// 启动任务生成器
	go func() {
		defer close(taskChan)
		generatePasswords(ctx, minLen, maxLen, taskChan, &found)
	}()

	// 启动进度报告
	progressDone := make(chan struct{})
	go func() {
		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				c := atomic.LoadInt64(&count)
				if atomic.LoadInt32(&found) == 0 {
					fmt.Printf("已尝试: %d 个密码...\n", c)
				}
			case <-ctx.Done():
				return
			case <-progressDone:
				return
			}
		}
	}()

	// 等待结果
	var password string
	select {
	case password = <-resultChan:
		// 找到密码，取消所有操作
		atomic.StoreInt32(&found, 1)
		cancel()
		close(progressDone)
		wg.Wait()
		return password
	case <-func() chan struct{} {
		done := make(chan struct{})
		go func() {
			wg.Wait()
			close(done)
		}()
		return done
	}():
		// 所有任务完成但没有找到密码
		atomic.StoreInt32(&found, 1)
		cancel()
		close(progressDone)
		return ""
	}
}

func worker(ctx context.Context, zipData []byte, taskChan <-chan string, resultChan chan<- string, wg *sync.WaitGroup, found *int32, count *int64) {
	defer wg.Done()

	for {
		select {
		case <-ctx.Done():
			return
		case password, ok := <-taskChan:
			if !ok {
				return
			}

			// 检查是否已经找到密码
			if atomic.LoadInt32(found) == 1 {
				return
			}

			atomic.AddInt64(count, 1)

			if testPassword(zipData, password) {
				// 尝试发送结果，如果已经找到则忽略
				if atomic.CompareAndSwapInt32(found, 0, 1) {
					select {
					case resultChan <- password:
					default:
					}
				}
				return
			}
		}
	}
}

func generatePasswords(ctx context.Context, minLen, maxLen int, taskChan chan<- string, found *int32) {
	// 计算总组合数（使用int64避免溢出）
	var totalCombinations int64
	for length := minLen; length <= maxLen; length++ {
		combos := int64(1)
		for i := 0; i < length; i++ {
			combos *= int64(len(charset))
		}
		totalCombinations += combos
	}

	fmt.Printf("预计总组合数: %d\n\n", totalCombinations)

	for length := minLen; length <= maxLen; length++ {
		if atomic.LoadInt32(found) == 1 {
			return
		}
		generatePasswordsOfLength(ctx, length, taskChan, found)
	}
}

func generatePasswordsOfLength(ctx context.Context, length int, taskChan chan<- string, found *int32) {
	indices := make([]int, length)

	for {
		// 检查是否已经找到密码或context已取消
		if atomic.LoadInt32(found) == 1 {
			return
		}

		select {
		case <-ctx.Done():
			return
		default:
		}

		// 生成密码
		password := make([]byte, length)
		for i := 0; i < length; i++ {
			password[i] = charset[indices[i]]
		}

		// 发送任务，如果找到密码则停止
		select {
		case taskChan <- string(password):
			// 成功发送任务
		case <-ctx.Done():
			return
		}

		// 递增索引
		if !incrementIndices(indices, len(charset)) {
			break
		}
	}
}

func incrementIndices(indices []int, base int) bool {
	for i := len(indices) - 1; i >= 0; i-- {
		indices[i]++
		if indices[i] < base {
			return true
		}
		indices[i] = 0
	}
	return false
}

func testPassword(zipData []byte, password string) bool {
	// 从内存中创建reader
	reader, err := zip.NewReader(bytes.NewReader(zipData), int64(len(zipData)))
	if err != nil {
		return false
	}

	// 尝试用密码解压第一个文件来验证密码
	if len(reader.File) == 0 {
		return false
	}

	file := reader.File[0]

	// 跳过未加密的文件（如果没有加密，直接返回false）
	// 注意：IsEncrypted()可能在某些版本中不存在，所以先尝试设置密码

	// 设置密码
	file.SetPassword(password)

	// 尝试打开文件
	rc, err := file.Open()
	if err != nil {
		// 如果打开失败，可能是密码错误或其他错误
		return false
	}
	defer rc.Close()

	// 尝试读取数据来验证密码
	// 对于加密的ZIP文件，如果密码正确，应该能读取数据
	// 如果密码错误，读取可能会失败或返回错误数据
	buf := make([]byte, 32)
	n, err := rc.Read(buf)

	// 如果能成功读取到数据，说明密码正确
	// 注意：某些ZIP实现可能在密码错误时也能读取但数据是乱码
	// 但通常alexmullins/zip会在密码错误时返回错误
	if err == nil || (err == io.EOF && n > 0) {
		return true
	}

	return false
}
