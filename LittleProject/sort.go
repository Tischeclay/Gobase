package main

import (
	"fmt"
	"time"
)

// ==================== 主函数 ====================

func main() {

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
