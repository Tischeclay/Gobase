package main

import (
	"fmt"
	"math/rand"
	"strings"
	"time"
)

// ==================== 基础归并排序 ====================

// MergeSort 标准归并排序（递归版）
func MergeSort(arr []int) []int {
	if len(arr) <= 1 {
		return arr
	}

	mid := len(arr) / 2
	left := MergeSort(arr[:mid])
	right := MergeSort(arr[mid:])

	return merge(left, right)
}

// merge 合并两个有序数组
func merge(left, right []int) []int {
	result := make([]int, 0, len(left)+len(right))
	i, j := 0, 0

	for i < len(left) && j < len(right) {
		if left[i] <= right[j] {
			result = append(result, left[i])
			i++
		} else {
			result = append(result, right[j])
			j++
		}
	}

	// 添加剩余元素
	result = append(result, left[i:]...)
	result = append(result, right[j:]...)

	return result
}

// ==================== 原地归并排序 ====================

// MergeSortInPlace 原地归并排序（不创建新数组）
func MergeSortInPlace(arr []int) {
	temp := make([]int, len(arr))
	mergeSortInPlace(arr, temp, 0, len(arr)-1)
}

func mergeSortInPlace(arr, temp []int, left, right int) {
	if left >= right {
		return
	}

	mid := left + (right-left)/2
	mergeSortInPlace(arr, temp, left, mid)
	mergeSortInPlace(arr, temp, mid+1, right)
	mergeInPlace(arr, temp, left, mid, right)
}

func mergeInPlace(arr, temp []int, left, mid, right int) {
	// 复制到临时数组
	for i := left; i <= right; i++ {
		temp[i] = arr[i]
	}

	i, j := left, mid+1
	for k := left; k <= right; k++ {
		if i > mid {
			arr[k] = temp[j]
			j++
		} else if j > right {
			arr[k] = temp[i]
			i++
		} else if temp[i] <= temp[j] {
			arr[k] = temp[i]
			i++
		} else {
			arr[k] = temp[j]
			j++
		}
	}
}

// ==================== 迭代归并排序 ====================

// MergeSortIterative 迭代归并排序（自底向上）
func MergeSortIterative(arr []int) {
	n := len(arr)
	temp := make([]int, n)

	// 逐步扩大合并区间
	for size := 1; size < n; size *= 2 {
		for left := 0; left < n-1; left += 2 * size {
			mid := left + size - 1
			right := min(left+2*size-1, n-1)

			if mid < right {
				mergeIterative(arr, temp, left, mid, right)
			}
		}
	}
}

func mergeIterative(arr, temp []int, left, mid, right int) {
	// 复制到临时数组
	for i := left; i <= right; i++ {
		temp[i] = arr[i]
	}

	i, j := left, mid+1
	for k := left; k <= right; k++ {
		if i > mid {
			arr[k] = temp[j]
			j++
		} else if j > right {
			arr[k] = temp[i]
			i++
		} else if temp[i] <= temp[j] {
			arr[k] = temp[i]
			i++
		} else {
			arr[k] = temp[j]
			j++
		}
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// ==================== 优化归并排序 ====================

// MergeSortOptimized 优化归并排序（小数组使用插入排序）
func MergeSortOptimized(arr []int) {
	if len(arr) <= 1 {
		return
	}
	temp := make([]int, len(arr))
	mergeSortOptimized(arr, temp, 0, len(arr)-1)
}

func mergeSortOptimized(arr, temp []int, left, right int) {
	// 小数组使用插入排序
	if right-left <= 15 {
		insertionSort(arr, left, right)
		return
	}

	mid := left + (right-left)/2
	mergeSortOptimized(arr, temp, left, mid)
	mergeSortOptimized(arr, temp, mid+1, right)

	// 如果已经有序，跳过合并
	if arr[mid] <= arr[mid+1] {
		return
	}

	mergeInPlace(arr, temp, left, mid, right)
}

// insertionSort 插入排序（用于小数组）
func insertionSort(arr []int, left, right int) {
	for i := left + 1; i <= right; i++ {
		key := arr[i]
		j := i - 1
		for j >= left && arr[j] > key {
			arr[j+1] = arr[j]
			j--
		}
		arr[j+1] = key
	}
}

// ==================== 并行归并排序 ====================

// MergeSortParallel 并行归并排序
func MergeSortParallel(arr []int) {
	if len(arr) <= 1 {
		return
	}
	temp := make([]int, len(arr))
	mergeSortParallel(arr, temp, 0, len(arr)-1)
}

func mergeSortParallel(arr, temp []int, left, right int) {
	if left >= right {
		return
	}

	// 小数组使用单线程
	if right-left <= 5000 {
		mergeSortInPlace(arr, temp, left, right)
		return
	}

	mid := left + (right-left)/2

	// 并行处理左右两部分
	done := make(chan bool)
	go func() {
		mergeSortParallel(arr, temp, left, mid)
		done <- true
	}()
	mergeSortParallel(arr, temp, mid+1, right)
	<-done

	// 合并
	if arr[mid] <= arr[mid+1] {
		return
	}
	mergeInPlace(arr, temp, left, mid, right)
}

// ==================== 外部归并排序（模拟） ====================

// ExternalMergeSort 外部归并排序（用于大数据）
func ExternalMergeSort(arr []int, chunkSize int) []int {
	if len(arr) <= chunkSize {
		sorted := make([]int, len(arr))
		copy(sorted, arr)
		MergeSortInPlace(sorted)
		return sorted
	}

	// 分块排序
	chunks := make([][]int, 0)
	for i := 0; i < len(arr); i += chunkSize {
		end := i + chunkSize
		if end > len(arr) {
			end = len(arr)
		}

		chunk := make([]int, end-i)
		copy(chunk, arr[i:end])
		MergeSortInPlace(chunk)
		chunks = append(chunks, chunk)
	}

	// 多路归并
	return multiWayMerge(chunks)
}

// multiWayMerge 多路归并
func multiWayMerge(chunks [][]int) []int {
	type Item struct {
		value int
		chunk int
		index int
	}

	// 创建最小堆
	heap := make([]Item, 0, len(chunks))
	for i, chunk := range chunks {
		if len(chunk) > 0 {
			heap = append(heap, Item{chunk[0], i, 0})
		}
	}

	// 构建堆
	for i := len(heap)/2 - 1; i >= 0; i-- {
		heapify(heap, i)
	}

	// 归并
	result := make([]int, 0)
	for len(heap) > 0 {
		// 取出最小值
		item := heap[0]
		result = append(result, item.value)

		// 从对应的块中取下一个元素
		if item.index+1 < len(chunks[item.chunk]) {
			heap[0] = Item{
				value: chunks[item.chunk][item.index+1],
				chunk: item.chunk,
				index: item.index + 1,
			}
			heapify(heap, 0)
		} else {
			// 移除该块
			heap[0] = heap[len(heap)-1]
			heap = heap[:len(heap)-1]
			if len(heap) > 0 {
				heapify(heap, 0)
			}
		}
	}

	return result
}

// ==================== 自然归并排序 ====================

// NaturalMergeSort 自然归并排序（利用已有有序序列）
func NaturalMergeSort(arr []int) {
	if len(arr) <= 1 {
		return
	}

	temp := make([]int, len(arr))
	for {
		// 查找所有有序段
		runs := findRuns(arr)
		if len(runs) <= 1 {
			break
		}

		// 合并相邻有序段
		newRuns := make([]int, 0)
		for i := 0; i < len(runs)-1; i += 2 {
			start := runs[i]
			mid := runs[i+1] - 1
			end := runs[i+2] - 1
			if i+2 < len(runs) {
				end = runs[i+2] - 1
			} else {
				end = len(arr) - 1
			}

			mergeInPlace(arr, temp, start, mid, end)
			newRuns = append(newRuns, start)
		}
		newRuns = append(newRuns, len(arr))

		// 更新runs
		runs = newRuns
	}
}

func findRuns(arr []int) []int {
	runs := []int{0}
	for i := 1; i < len(arr); i++ {
		if arr[i-1] > arr[i] {
			runs = append(runs, i)
		}
	}
	runs = append(runs, len(arr))
	return runs
}

// ==================== 辅助函数 ====================

// GenerateRandomArray 生成随机数组
func GenerateRandomArray(n int) []int {
	arr := make([]int, n)
	for i := 0; i < n; i++ {
		arr[i] = rand.Intn(n * 10)
	}
	return arr
}

// GenerateSortedArray 生成已排序数组
func GenerateSortedArray(n int) []int {
	arr := make([]int, n)
	for i := 0; i < n; i++ {
		arr[i] = i
	}
	return arr
}

// GenerateReverseArray 生成逆序数组
func GenerateReverseArray(n int) []int {
	arr := make([]int, n)
	for i := 0; i < n; i++ {
		arr[i] = n - i
	}
	return arr
}

// IsSorted 检查数组是否已排序
func IsSorted(arr []int) bool {
	for i := 1; i < len(arr); i++ {
		if arr[i-1] > arr[i] {
			return false
		}
	}
	return true
}

// CopyArray 复制数组
func CopyArray(arr []int) []int {
	copyArr := make([]int, len(arr))
	copy(copyArr, arr)
	return copyArr
}

// PrintArray 打印数组（限制长度）
func PrintArray(arr []int, limit int) {
	if len(arr) <= limit {
		fmt.Println(arr)
	} else {
		fmt.Printf("%v... (长度: %d)\n", arr[:limit], len(arr))
	}
}

// ==================== 性能测试 ====================

func benchmarkMergeSort(name string, sortFunc func([]int), arr []int) time.Duration {
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

func runBenchmarks(size int) {
	fmt.Printf("\n========== 归并排序性能测试 (数组大小: %d) ==========\n", size)

	testCases := []struct {
		name string
		gen  func(int) []int
	}{
		{"随机数组", GenerateRandomArray},
		{"已排序数组", GenerateSortedArray},
		{"逆序数组", GenerateReverseArray},
	}

	algorithms := []struct {
		name string
		fn   func([]int)
	}{
		{"标准归并排序", func(arr []int) { arr2 := MergeSort(arr); copy(arr, arr2) }},
		{"原地归并排序", MergeSortInPlace},
		{"迭代归并排序", MergeSortIterative},
		{"优化归并排序", MergeSortOptimized},
		{"并行归并排序", MergeSortParallel},
	}

	for _, tc := range testCases {
		fmt.Printf("\n📊 %s:\n", tc.name)
		arr := tc.gen(size)

		for _, algo := range algorithms {
			duration := benchmarkMergeSort(algo.name, algo.fn, arr)
			if duration > 0 {
				fmt.Printf("   %-15s: %v\n", algo.name, duration)
			}
		}
	}
}

// ==================== 可视化演示 ====================

func visualMergeSort() {
	fmt.Println("\n========== 归并排序可视化演示 ==========")

	arr := []int{38, 27, 43, 3, 9, 82, 10}
	fmt.Printf("原始数组: %v\n", arr)
	fmt.Println("\n归并排序过程:")

	result := mergeSortVisual(arr, 0)
	fmt.Printf("\n排序结果: %v\n", result)
}

func mergeSortVisual(arr []int, depth int) []int {
	if len(arr) <= 1 {
		indent := strings.Repeat("  ", depth)
		fmt.Printf("%s返回: %v\n", indent, arr)
		return arr
	}

	mid := len(arr) / 2
	indent := strings.Repeat("  ", depth)

	fmt.Printf("%s分割: %v -> %v | %v\n", indent, arr, arr[:mid], arr[mid:])

	left := mergeSortVisual(arr[:mid], depth+1)
	right := mergeSortVisual(arr[mid:], depth+1)

	merged := mergeVisual(left, right, depth)
	fmt.Printf("%s合并: %v + %v = %v\n", indent, left, right, merged)

	return merged
}

func mergeVisual(left, right []int, depth int) []int {
	result := make([]int, 0, len(left)+len(right))
	i, j := 0, 0

	for i < len(left) && j < len(right) {
		if left[i] <= right[j] {
			result = append(result, left[i])
			i++
		} else {
			result = append(result, right[j])
			j++
		}
	}

	result = append(result, left[i:]...)
	result = append(result, right[j:]...)

	return result
}

// ==================== 算法复杂度分析 ====================

func printComplexity() {
	fmt.Println("\n========== 归并排序复杂度分析 ==========")
	fmt.Println(`
┌─────────────────────────────────────────────────────────────────────┐
│                    归并排序算法复杂度                                 │
├─────────────────────────────────────────────────────────────────────┤
│                                                                      │
│  时间复杂度:                                                         │
│    • 最好情况: O(n log n)                                            │
│    • 最坏情况: O(n log n)                                            │
│    • 平均情况: O(n log n)                                            │
│                                                                      │
│  空间复杂度:                                                         │
│    • 标准归并: O(n)                                                  │
│    • 原地归并: O(n)                                                  │
│                                                                      │
│  稳定性: 稳定                                                        │
│                                                                      │
│  优点:                                                               │
│    • 时间复杂度稳定，不受输入影响                                     │
│    • 稳定排序                                                        │
│    • 适合链表排序                                                    │
│    • 适合外部排序                                                    │
│                                                                      │
│  缺点:                                                               │
│    • 需要额外空间 O(n)                                               │
│    • 小数组效率不如插入排序                                           │
│                                                                      │
└─────────────────────────────────────────────────────────────────────┘
    `)
}

// ==================== 链表归并排序 ====================

// ListNode 链表节点
type ListNode struct {
	Val  int
	Next *ListNode
}

// MergeSortList 链表归并排序
func MergeSortList(head *ListNode) *ListNode {
	if head == nil || head.Next == nil {
		return head
	}

	// 找到中间节点
	mid := findMiddle(head)
	rightHead := mid.Next
	mid.Next = nil

	// 递归排序
	left := MergeSortList(head)
	right := MergeSortList(rightHead)

	// 合并
	return mergeList(left, right)
}

func findMiddle(head *ListNode) *ListNode {
	slow, fast := head, head
	for fast.Next != nil && fast.Next.Next != nil {
		slow = slow.Next
		fast = fast.Next.Next
	}
	return slow
}

func mergeList(l1, l2 *ListNode) *ListNode {
	dummy := &ListNode{}
	current := dummy

	for l1 != nil && l2 != nil {
		if l1.Val <= l2.Val {
			current.Next = l1
			l1 = l1.Next
		} else {
			current.Next = l2
			l2 = l2.Next
		}
		current = current.Next
	}

	if l1 != nil {
		current.Next = l1
	}
	if l2 != nil {
		current.Next = l2
	}

	return dummy.Next
}

func printList(head *ListNode) {
	for head != nil {
		fmt.Printf("%d", head.Val)
		if head.Next != nil {
			fmt.Print(" -> ")
		}
		head = head.Next
	}
	fmt.Println()
}

// ==================== 主函数 ====================

func main() {
	rand.Seed(time.Now().UnixNano())

	fmt.Println("==================== Go语言归并排序完整实现 ====================")

	// 算法复杂度分析
	printComplexity()

	// 可视化演示
	visualMergeSort()

	// 链表归并排序
	fmt.Println("\n========== 链表归并排序 ==========")
	head := &ListNode{Val: 4}
	head.Next = &ListNode{Val: 2}
	head.Next.Next = &ListNode{Val: 1}
	head.Next.Next.Next = &ListNode{Val: 3}

	fmt.Print("原始链表: ")
	printList(head)

	sorted := MergeSortList(head)
	fmt.Print("排序后: ")
	printList(sorted)

	// 性能测试
	runBenchmarks(1000)
	runBenchmarks(10000)
	runBenchmarks(100000)

	// 外部归并排序测试
	fmt.Println("\n========== 外部归并排序测试 ==========")
	largeArr := GenerateRandomArray(10000)
	fmt.Printf("原始数组大小: %d\n", len(largeArr))

	start := time.Now()
	result := ExternalMergeSort(largeArr, 1000)
	fmt.Printf("外部归并排序耗时: %v\n", time.Since(start))

	if IsSorted(result) {
		fmt.Println("✅ 排序成功")
	}

	fmt.Println("\n✅ 所有测试完成!")
}
