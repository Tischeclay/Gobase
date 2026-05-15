package main

import (
	"fmt"
	"time"
)

// ==================== 主函数 ====================

func main() {

	// 并发排序测试
	fmt.Println("\n========== 并发排序测试 ==========")
	size := 50000
	arr := GenerateRandomArray(size)

	fmt.Printf("并发快速排序 (数组大小: %d)\n", size)
	start := time.Now()

	if IsSorted(arr) {
		fmt.Printf("✅ 排序成功，耗时: %v\n", duration)
	} else {
		fmt.Println("❌ 排序失败")
	}
}
