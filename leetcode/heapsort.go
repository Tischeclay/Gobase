package main

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
