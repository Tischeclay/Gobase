package main

import (
	"fmt"
	"math/rand"
	"time"
)

// ==================== 并发排序 ====================

// ConcurrentQuickSort 并发快速排序（使用goroutine）
func ConcurrentQuickSort(arr []int) {
	if len(arr) <= 1 {
		return
	}

	done := make(chan bool)
	go concurrentQuickSort(arr, 0, len(arr)-1, done)
	<-done
}

func concurrentQuickSort(arr []int, low, high int, done chan bool) {
	defer func() { done <- true }()

	if low >= high {
		return
	}

	// 小数组使用普通快排
	if high-low < 1000 {
		quickSortRecursive(arr, low, high)
		return
	}

	// 分区
	pi := partition(arr, low, high)

	// 并发排序左右部分
	leftDone := make(chan bool)
	rightDone := make(chan bool)

	go concurrentQuickSort(arr, low, pi-1, leftDone)
	go concurrentQuickSort(arr, pi+1, high, rightDone)

	<-leftDone
	<-rightDone
}

// ==================== 主函数 ====================

func main() {
	rand.Seed(time.Now().UnixNano())

	// 算法说明
	ExplainAlgorithms()

	// 演示排序过程
	DemoSort()

	// 小规模性能测试
	RunBenchmarks(1000)

	// 中等规模性能测试
	RunBenchmarks(10000)

	// 大规模性能测试（可选）
	fmt.Print("\n是否进行大规模性能测试 (100000)? (y/n): ")
	var input string
	fmt.Scanln(&input)
	if input == "y" || input == "Y" {
		RunBenchmarks(100000)
	}

	// 并发排序测试
	fmt.Println("\n========== 并发排序测试 ==========")
	size := 50000
	arr := GenerateRandomArray(size)

	fmt.Printf("并发快速排序 (数组大小: %d)\n", size)
	start := time.Now()
	ConcurrentQuickSort(arr)
	duration := time.Since(start)

	if IsSorted(arr) {
		fmt.Printf("✅ 排序成功，耗时: %v\n", duration)
	} else {
		fmt.Println("❌ 排序失败")
	}
}
