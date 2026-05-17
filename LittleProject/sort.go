package main

import (
	"fmt"
)

// ==================== 主函数 ====================

func main() {

	// 并发排序测试
	fmt.Println("\n========== 并发排序测试 ==========")
	size := 50000
	arr := GenerateRandomArray(size)

	if IsSorted(arr) {
		fmt.Printf("✅ 排序成功，耗时: %v\n", duration)
	} else {
		fmt.Println("❌ 排序失败")
	}
}
