package main

import (
	"fmt"
	"math/rand"
	"time"
)

// CopyArray 复制数组
func CopyArray(arr []int) []int {
	copyArr := make([]int, len(arr))
	copy(copyArr, arr)
	return copyArr
}

// 打印数组（限制长度）
func PrintArray(arr []int, limit int) {
	if len(arr) <= limit {
		fmt.Println(arr)
	} else {
		fmt.Printf("%v... (长度: %d)\n", arr[:limit], len(arr))
	}
}

// ==================== 性能测试 ====================

// SortBenchmark 排序性能测试
func SortBenchmark(name string, sortFunc func([]int), arr []int) time.Duration {
	testArr := CopyArray(arr)
	start := time.Now()
	sortFunc(testArr)
	duration := time.Since(start)

	if !IsSorted(testArr) {
		fmt.Printf("❌ %s 排序失败！\n", name)
		return 0
	}

	return duration
}

// RunBenchmarks 运行所有基准测试
func RunBenchmarks(size int) {
	fmt.Printf("\n========== 排序性能测试 (数组大小: %d) ==========\n", size)
	fmt.Println()

	// 测试不同场景
	testCases := []struct {
		name string
		gen  func(int) []int
	}{
		{"随机数组", GenerateRandomArray},
		{"已排序数组", GenerateSortedArray},
		{"逆序数组", GenerateReverseArray},
		{"重复元素数组", GenerateDuplicateArray},
	}

	// 要测试的排序算法
	algorithms := []struct {
		name string
		fn   func([]int)
	}{
		{"冒泡排序", BubbleSort},
		{"优化冒泡排序", BubbleSortOptimized},
		{"快速排序", QuickSort},
		{"随机基准快排", QuickSortRandom},
		{"三路快排", QuickSortThreeWay},
	}

	for _, tc := range testCases {
		fmt.Printf("📊 %s:\n", tc.name)
		arr := tc.gen(size)

		for _, algo := range algorithms {
			duration := SortBenchmark(algo.name, algo.fn, arr)
			if duration > 0 {
				fmt.Printf("   %-15s: %v\n", algo.name, duration)
			}
		}
		fmt.Println()
	}
}

// ==================== 交互式演示 ====================

// DemoSort 演示排序过程
func DemoSort() {
	fmt.Println("========== 排序算法演示 ==========\n")

	// 创建测试数组
	arr := []int{64, 34, 25, 12, 22, 11, 90, 5, 77, 33}

	fmt.Printf("原始数组: %v\n\n", arr)

	// 冒泡排序
	arr1 := CopyArray(arr)
	fmt.Println("1. 冒泡排序:")
	fmt.Printf("   排序前: %v\n", arr1)
	BubbleSort(arr1)
	fmt.Printf("   排序后: %v\n\n", arr1)

	// 快速排序
	arr2 := CopyArray(arr)
	fmt.Println("2. 快速排序:")
	fmt.Printf("   排序前: %v\n", arr2)
	QuickSort(arr2)
	fmt.Printf("   排序后: %v\n\n", arr2)

	// 优化冒泡排序
	arr3 := CopyArray(arr)
	fmt.Println("3. 优化冒泡排序:")
	fmt.Printf("   排序前: %v\n", arr3)
	BubbleSortOptimized(arr3)
	fmt.Printf("   排序后: %v\n\n", arr3)

	// 随机基准快排
	arr4 := CopyArray(arr)
	fmt.Println("4. 随机基准快排:")
	fmt.Printf("   排序前: %v\n", arr4)
	QuickSortRandom(arr4)
	fmt.Printf("   排序后: %v\n\n", arr4)

	// 三路快排
	arr5 := []int{3, 1, 4, 1, 5, 9, 2, 6, 5, 3, 5}
	fmt.Println("5. 三路快排 (处理重复元素):")
	fmt.Printf("   排序前: %v\n", arr5)
	QuickSortThreeWay(arr5)
	fmt.Printf("   排序后: %v\n\n", arr5)
}

// ==================== 算法详解 ====================

func ExplainAlgorithms() {
	fmt.Println("========== 算法复杂度分析 ==========\n")

	fmt.Println("📚 冒泡排序 (Bubble Sort):")
	fmt.Println("   原理: 重复遍历数组，比较相邻元素并交换")
	fmt.Println("   时间复杂度:")
	fmt.Println("     - 最好: O(n) (已排序)")
	fmt.Println("     - 平均: O(n²)")
	fmt.Println("     - 最坏: O(n²) (逆序)")
	fmt.Println("   空间复杂度: O(1)")
	fmt.Println("   稳定性: 稳定")
	fmt.Println()

	fmt.Println("📚 快速排序 (Quick Sort):")
	fmt.Println("   原理: 分治策略，选择一个基准，将小于基准的放左边，大于的放右边")
	fmt.Println("   时间复杂度:")
	fmt.Println("     - 最好: O(n log n)")
	fmt.Println("     - 平均: O(n log n)")
	fmt.Println("     - 最坏: O(n²) (已排序或逆序)")
	fmt.Println("   空间复杂度: O(log n) (递归栈)")
	fmt.Println("   稳定性: 不稳定")
	fmt.Println()

	fmt.Println("✨ 优化技巧:")
	fmt.Println("   1. 随机基准: 避免最坏情况")
	fmt.Println("   2. 三路快排: 处理大量重复元素")
	fmt.Println("   3. 小数组使用插入排序")
	fmt.Println("   4. 尾递归优化")
	fmt.Println()
}

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
