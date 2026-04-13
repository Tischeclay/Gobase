package main

import (
	"fmt"
	"math/rand"
	"time"
)

// ==================== 基础堆排序 ====================

// HeapSort 标准堆排序
func HeapSort(arr []int) {
	n := len(arr)

	// 构建最大堆
	for i := n/2 - 1; i >= 0; i-- {
		heapify(arr, n, i)
	}

	// 一个个从堆顶取出元素
	for i := n - 1; i > 0; i-- {
		// 将当前根节点（最大值）与末尾元素交换
		arr[0], arr[i] = arr[i], arr[0]
		// 重新调整堆
		heapify(arr, i, 0)
	}
}

// heapify 调整堆结构
func heapify(arr []int, n, i int) {
	largest := i     // 初始化最大值为根节点
	left := 2*i + 1  // 左子节点
	right := 2*i + 2 // 右子节点

	// 如果左子节点存在且大于根节点
	if left < n && arr[left] > arr[largest] {
		largest = left
	}

	// 如果右子节点存在且大于当前最大值
	if right < n && arr[right] > arr[largest] {
		largest = right
	}

	// 如果最大值不是根节点
	if largest != i {
		arr[i], arr[largest] = arr[largest], arr[i]
		// 递归调整受影响的子树
		heapify(arr, n, largest)
	}
}

// ==================== 迭代版堆排序 ====================

// HeapSortIterative 迭代版堆排序（避免递归）
func HeapSortIterative(arr []int) {
	n := len(arr)

	// 构建最大堆（迭代方式）
	for i := n/2 - 1; i >= 0; i-- {
		heapifyIterative(arr, n, i)
	}

	// 提取元素
	for i := n - 1; i > 0; i-- {
		arr[0], arr[i] = arr[i], arr[0]
		heapifyIterative(arr, i, 0)
	}
}

// heapifyIterative 迭代版堆调整
func heapifyIterative(arr []int, n, i int) {
	for {
		largest := i
		left := 2*i + 1
		right := 2*i + 2

		if left < n && arr[left] > arr[largest] {
			largest = left
		}
		if right < n && arr[right] > arr[largest] {
			largest = right
		}

		if largest == i {
			break
		}

		arr[i], arr[largest] = arr[largest], arr[i]
		i = largest
	}
}

// ==================== 最小堆排序 ====================

// MinHeapSort 最小堆排序（降序）
func MinHeapSort(arr []int) {
	n := len(arr)

	// 构建最小堆
	for i := n/2 - 1; i >= 0; i-- {
		minHeapify(arr, n, i)
	}

	// 提取元素
	for i := n - 1; i > 0; i-- {
		arr[0], arr[i] = arr[i], arr[0]
		minHeapify(arr, i, 0)
	}
}

// minHeapify 最小堆调整
func minHeapify(arr []int, n, i int) {
	smallest := i
	left := 2*i + 1
	right := 2*i + 2

	if left < n && arr[left] < arr[smallest] {
		smallest = left
	}
	if right < n && arr[right] < arr[smallest] {
		smallest = right
	}

	if smallest != i {
		arr[i], arr[smallest] = arr[smallest], arr[i]
		minHeapify(arr, n, smallest)
	}
}

// ==================== 堆排序变体 ====================

// HeapSortWithK 只找出前K个最大元素
func HeapSortWithK(arr []int, k int) []int {
	if k > len(arr) {
		k = len(arr)
	}

	// 构建最小堆（用于找出前K个最大元素）
	heap := make([]int, k)
	copy(heap, arr[:k])

	// 构建最小堆
	for i := k/2 - 1; i >= 0; i-- {
		minHeapify(heap, k, i)
	}

	// 处理剩余元素
	for i := k; i < len(arr); i++ {
		if arr[i] > heap[0] {
			heap[0] = arr[i]
			minHeapify(heap, k, 0)
		}
	}

	// 对结果排序
	result := make([]int, k)
	copy(result, heap)
	HeapSort(result)

	return result
}

// ==================== 并发堆排序 ====================

// ParallelHeapSort 并发堆排序（分块处理）
func ParallelHeapSort(arr []int, numWorkers int) {
	if len(arr) <= 1000 {
		HeapSort(arr)
		return
	}

	// 分块
	chunkSize := len(arr) / numWorkers
	chunks := make([][]int, numWorkers)

	for i := 0; i < numWorkers; i++ {
		start := i * chunkSize
		end := start + chunkSize
		if i == numWorkers-1 {
			end = len(arr)
		}

		chunk := make([]int, end-start)
		copy(chunk, arr[start:end])
		chunks[i] = chunk
	}

	// 并发排序各块
	done := make(chan bool, numWorkers)
	for i := 0; i < numWorkers; i++ {
		go func(idx int) {
			HeapSort(chunks[idx])
			done <- true
		}(i)
	}

	// 等待所有块排序完成
	for i := 0; i < numWorkers; i++ {
		<-done
	}

	// 合并排序结果（使用堆合并）
	mergeSortedChunks(arr, chunks)
}

func mergeSortedChunks(arr []int, chunks [][]int) {
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
		heapifyItems(heap, i)
	}

	// 合并
	idx := 0
	for len(heap) > 0 {
		// 取出最小值
		item := heap[0]
		arr[idx] = item.value
		idx++

		// 从对应的块中取下一个元素
		if item.index+1 < len(chunks[item.chunk]) {
			heap[0] = Item{
				value: chunks[item.chunk][item.index+1],
				chunk: item.chunk,
				index: item.index + 1,
			}
			heapifyItems(heap, 0)
		} else {
			// 移除该块
			heap[0] = heap[len(heap)-1]
			heap = heap[:len(heap)-1]
			if len(heap) > 0 {
				heapifyItems(heap, 0)
			}
		}
	}
}

func heapifyItems(heap []Item, i int) {
	n := len(heap)
	smallest := i
	left := 2*i + 1
	right := 2*i + 2

	if left < n && heap[left].value < heap[smallest].value {
		smallest = left
	}
	if right < n && heap[right].value < heap[smallest].value {
		smallest = right
	}

	if smallest != i {
		heap[i], heap[smallest] = heap[smallest], heap[i]
		heapifyItems(heap, smallest)
	}
}

// ==================== 堆数据结构实现 ====================

// MaxHeap 最大堆数据结构
type MaxHeap struct {
	data []int
}

func NewMaxHeap() *MaxHeap {
	return &MaxHeap{
		data: make([]int, 0),
	}
}

func (h *MaxHeap) Push(value int) {
	h.data = append(h.data, value)
	h.up(len(h.data) - 1)
}

func (h *MaxHeap) Pop() int {
	if len(h.data) == 0 {
		return -1
	}

	max := h.data[0]
	h.data[0] = h.data[len(h.data)-1]
	h.data = h.data[:len(h.data)-1]
	h.down(0)

	return max
}

func (h *MaxHeap) Peek() int {
	if len(h.data) == 0 {
		return -1
	}
	return h.data[0]
}

func (h *MaxHeap) Size() int {
	return len(h.data)
}

func (h *MaxHeap) up(i int) {
	for i > 0 {
		parent := (i - 1) / 2
		if h.data[parent] >= h.data[i] {
			break
		}
		h.data[parent], h.data[i] = h.data[i], h.data[parent]
		i = parent
	}
}

func (h *MaxHeap) down(i int) {
	n := len(h.data)
	for {
		largest := i
		left := 2*i + 1
		right := 2*i + 2

		if left < n && h.data[left] > h.data[largest] {
			largest = left
		}
		if right < n && h.data[right] > h.data[largest] {
			largest = right
		}

		if largest == i {
			break
		}

		h.data[i], h.data[largest] = h.data[largest], h.data[i]
		i = largest
	}
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

func benchmarkHeapSort(name string, sortFunc func([]int), arr []int) time.Duration {
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
	fmt.Printf("\n========== 堆排序性能测试 (数组大小: %d) ==========\n", size)

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
		{"标准堆排序", HeapSort},
		{"迭代堆排序", HeapSortIterative},
		{"原地堆排序", InPlaceHeapSort},
		{"最小堆排序", MinHeapSort},
	}

	for _, tc := range testCases {
		fmt.Printf("\n📊 %s:\n", tc.name)
		arr := tc.gen(size)

		for _, algo := range algorithms {
			duration := benchmarkHeapSort(algo.name, algo.fn, arr)
			if duration > 0 {
				fmt.Printf("   %-15s: %v\n", algo.name, duration)
			}
		}
	}
}

// ==================== 堆数据结构演示 ====================

func demoHeap() {
	fmt.Println("\n========== 堆数据结构演示 ==========")

	heap := NewMaxHeap()

	// 插入元素
	values := []int{3, 1, 4, 1, 5, 9, 2, 6, 5, 3, 5}
	fmt.Printf("插入元素: %v\n", values)

	for _, v := range values {
		heap.Push(v)
		fmt.Printf("插入 %d 后，堆顶: %d\n", v, heap.Peek())
	}

	fmt.Printf("\n堆大小: %d\n", heap.Size())

	// 弹出元素
	fmt.Println("\n弹出所有元素:")
	for heap.Size() > 0 {
		fmt.Printf("%d ", heap.Pop())
	}
	fmt.Println()
}

// ==================== 可视化堆排序 ====================

func visualHeapSort(arr []int) {
	fmt.Printf("\n原始数组: %v\n", arr)

	n := len(arr)

	// 构建堆
	fmt.Println("\n构建最大堆过程:")
	for i := n/2 - 1; i >= 0; i-- {
		heapifyWithPrint(arr, n, i)
		fmt.Printf("  调整节点 %d 后: %v\n", i, arr)
	}

	fmt.Printf("\n最大堆: %v\n", arr)

	// 排序
	fmt.Println("\n排序过程:")
	for i := n - 1; i > 0; i-- {
		fmt.Printf("  交换堆顶 %d 和末尾 %d\n", arr[0], arr[i])
		arr[0], arr[i] = arr[i], arr[0]
		fmt.Printf("  交换后: %v\n", arr)
		heapifyWithPrint(arr, i, 0)
		fmt.Printf("  调整后: %v\n", arr)
	}

	fmt.Printf("\n排序结果: %v\n", arr)
}

func heapifyWithPrint(arr []int, n, i int) {
	largest := i
	left := 2*i + 1
	right := 2*i + 2

	if left < n && arr[left] > arr[largest] {
		largest = left
	}
	if right < n && arr[right] > arr[largest] {
		largest = right
	}

	if largest != i {
		fmt.Printf("    交换 %d 和 %d\n", arr[i], arr[largest])
		arr[i], arr[largest] = arr[largest], arr[i]
		heapifyWithPrint(arr, n, largest)
	}
}

// ==================== 主函数 ====================

func main() {
	rand.Seed(time.Now().UnixNano())

	// 算法详解
	explainHeapSort()

	// 可视化演示
	fmt.Println("\n========== 可视化堆排序演示 ==========")
	demoArray := []int{4, 10, 3, 5, 1, 2, 8, 7, 6, 9}
	visualHeapSort(CopyArray(demoArray))

	// 堆数据结构演示
	demoHeap()

	// 前K个最大元素
	fmt.Println("\n========== 找出前K个最大元素 ==========")
	arr := GenerateRandomArray(20)
	fmt.Printf("原始数组: %v\n", arr)
	k := 5
	topK := HeapSortWithK(arr, k)
	fmt.Printf("前 %d 个最大元素: %v\n", k, topK)

	// 性能测试
	runBenchmarks(1000)
	runBenchmarks(10000)
	runBenchmarks(100000)

	// 并发堆排序测试
	//fmt.Println("\n========== 并发堆排序测试 ==========")
	largeArr := GenerateRandomArray(100000)
	fmt.Printf("数组大小: %d\n", len(largeArr))

	start := time.Now()
	ParallelHeapSort(CopyArray(largeArr), 4)
	fmt.Printf("并发堆排序耗时: %v\n", time.Since(start))

	start = time.Now()
	HeapSort(CopyArray(largeArr))
	fmt.Printf("标准堆排序耗时: %v\n", time.Since(start))

	fmt.Println("\n✅ 所有测试完成!")
}
