package main

import (
	"fmt"
	"strings"
)

// ==================== 单链表节点 ====================

// ListNode 单链表节点
type ListNode struct {
	Val  int
	Next *ListNode
}

// LinkedList 单链表
type LinkedList struct {
	Head *ListNode
	Size int
}

// ==================== 单链表基本操作 ====================

// NewLinkedList 创建新链表
func NewLinkedList() *LinkedList {
	return &LinkedList{
		Head: nil,
		Size: 0,
	}
}

// AddAtHead 在头部添加节点
func (l *LinkedList) AddAtHead(val int) {
	node := &ListNode{Val: val, Next: l.Head}
	l.Head = node
	l.Size++
}

// AddAtTail 在尾部添加节点
func (l *LinkedList) AddAtTail(val int) {
	node := &ListNode{Val: val}
	l.Size++

	if l.Head == nil {
		l.Head = node
		return
	}

	current := l.Head
	for current.Next != nil {
		current = current.Next
	}
	current.Next = node
}

// AddAtIndex 在指定位置添加节点
func (l *LinkedList) AddAtIndex(index, val int) bool {
	if index < 0 || index > l.Size {
		return false
	}

	if index == 0 {
		l.AddAtHead(val)
		return true
	}

	node := &ListNode{Val: val}
	current := l.Head
	for i := 0; i < index-1; i++ {
		current = current.Next
	}

	node.Next = current.Next
	current.Next = node
	l.Size++

	return true
}

// Get 获取指定位置的节点值
func (l *LinkedList) Get(index int) (int, bool) {
	if index < 0 || index >= l.Size {
		return 0, false
	}

	current := l.Head
	for i := 0; i < index; i++ {
		current = current.Next
	}

	return current.Val, true
}

// DeleteAtIndex 删除指定位置的节点
func (l *LinkedList) DeleteAtIndex(index int) bool {
	if index < 0 || index >= l.Size {
		return false
	}

	if index == 0 {
		l.Head = l.Head.Next
		l.Size--
		return true
	}

	current := l.Head
	for i := 0; i < index-1; i++ {
		current = current.Next
	}

	current.Next = current.Next.Next
	l.Size--

	return true
}

// DeleteByValue 删除第一个值为val的节点
func (l *LinkedList) DeleteByValue(val int) bool {
	if l.Head == nil {
		return false
	}

	if l.Head.Val == val {
		l.Head = l.Head.Next
		l.Size--
		return true
	}

	current := l.Head
	for current.Next != nil && current.Next.Val != val {
		current = current.Next
	}

	if current.Next != nil {
		current.Next = current.Next.Next
		l.Size--
		return true
	}

	return false
}

// Find 查找值为val的节点位置
func (l *LinkedList) Find(val int) int {
	current := l.Head
	index := 0
	for current != nil {
		if current.Val == val {
			return index
		}
		current = current.Next
		index++
	}
	return -1
}

// Update 更新指定位置的值
func (l *LinkedList) Update(index, val int) bool {
	if index < 0 || index >= l.Size {
		return false
	}

	current := l.Head
	for i := 0; i < index; i++ {
		current = current.Next
	}
	current.Val = val

	return true
}

// ToSlice 转换为切片
func (l *LinkedList) ToSlice() []int {
	result := make([]int, 0, l.Size)
	current := l.Head
	for current != nil {
		result = append(result, current.Val)
		current = current.Next
	}
	return result
}

// Print 打印链表
func (l *LinkedList) Print() {
	elements := make([]string, 0)
	current := l.Head
	for current != nil {
		elements = append(elements, fmt.Sprintf("%d", current.Val))
		current = current.Next
	}
	fmt.Printf("链表: %s (长度: %d)\n", strings.Join(elements, " -> "), l.Size)
}

// IsEmpty 判断链表是否为空
func (l *LinkedList) IsEmpty() bool {
	return l.Size == 0
}

// GetSize 获取链表长度
func (l *LinkedList) GetSize() int {
	return l.Size
}

// ==================== 单链表高级操作 ====================

// Reverse 反转链表
func (l *LinkedList) Reverse() {
	var prev *ListNode
	current := l.Head

	for current != nil {
		next := current.Next
		current.Next = prev
		prev = current
		current = next
	}

	l.Head = prev
}

// HasCycle 检测环
func (l *LinkedList) HasCycle() bool {
	if l.Head == nil {
		return false
	}

	slow := l.Head
	fast := l.Head

	for fast != nil && fast.Next != nil {
		slow = slow.Next
		fast = fast.Next.Next
		if slow == fast {
			return true
		}
	}

	return false
}

// GetMiddle 获取中间节点
func (l *LinkedList) GetMiddle() *ListNode {
	if l.Head == nil {
		return nil
	}

	slow := l.Head
	fast := l.Head

	for fast != nil && fast.Next != nil {
		slow = slow.Next
		fast = fast.Next.Next
	}

	return slow
}

// Merge 合并两个有序链表
func Merge(l1, l2 *LinkedList) *LinkedList {
	dummy := &ListNode{}
	current := dummy

	p1 := l1.Head
	p2 := l2.Head

	for p1 != nil && p2 != nil {
		if p1.Val < p2.Val {
			current.Next = p1
			p1 = p1.Next
		} else {
			current.Next = p2
			p2 = p2.Next
		}
		current = current.Next
	}

	if p1 != nil {
		current.Next = p1
	}
	if p2 != nil {
		current.Next = p2
	}

	result := NewLinkedList()
	result.Head = dummy.Next
	result.Size = l1.Size + l2.Size

	return result
}

// RemoveNthFromEnd 删除倒数第n个节点
func (l *LinkedList) RemoveNthFromEnd(n int) bool {
	if n <= 0 || n > l.Size {
		return false
	}

	dummy := &ListNode{Next: l.Head}
	fast := dummy
	slow := dummy

	// fast先走n步
	for i := 0; i < n; i++ {
		fast = fast.Next
	}

	// 同时移动
	for fast.Next != nil {
		slow = slow.Next
		fast = fast.Next
	}

	// 删除节点
	slow.Next = slow.Next.Next
	l.Head = dummy.Next
	l.Size--

	return true
}

// ==================== 双链表 ====================

// DoubleListNode 双链表节点
type DoubleListNode struct {
	Val  int
	Prev *DoubleListNode
	Next *DoubleListNode
}

// DoublyLinkedList 双链表
type DoublyLinkedList struct {
	Head *DoubleListNode
	Tail *DoubleListNode
	Size int
}

// NewDoublyLinkedList 创建双链表
func NewDoublyLinkedList() *DoublyLinkedList {
	return &DoublyLinkedList{
		Head: nil,
		Tail: nil,
		Size: 0,
	}
}

// AddAtHead 在头部添加
func (l *DoublyLinkedList) AddAtHead(val int) {
	node := &DoubleListNode{Val: val}

	if l.Head == nil {
		l.Head = node
		l.Tail = node
	} else {
		node.Next = l.Head
		l.Head.Prev = node
		l.Head = node
	}
	l.Size++
}

// AddAtTail 在尾部添加
func (l *DoublyLinkedList) AddAtTail(val int) {
	node := &DoubleListNode{Val: val}

	if l.Tail == nil {
		l.Head = node
		l.Tail = node
	} else {
		node.Prev = l.Tail
		l.Tail.Next = node
		l.Tail = node
	}
	l.Size++
}

// DeleteNode 删除节点
func (l *DoublyLinkedList) DeleteNode(node *DoubleListNode) {
	if node == nil {
		return
	}

	if node.Prev != nil {
		node.Prev.Next = node.Next
	} else {
		l.Head = node.Next
	}

	if node.Next != nil {
		node.Next.Prev = node.Prev
	} else {
		l.Tail = node.Prev
	}

	l.Size--
}

// DeleteByValue 删除第一个值为val的节点
func (l *DoublyLinkedList) DeleteByValue(val int) bool {
	current := l.Head
	for current != nil {
		if current.Val == val {
			l.DeleteNode(current)
			return true
		}
		current = current.Next
	}
	return false
}

// PrintForward 正向打印
func (l *DoublyLinkedList) PrintForward() {
	elements := make([]string, 0)
	current := l.Head
	for current != nil {
		elements = append(elements, fmt.Sprintf("%d", current.Val))
		current = current.Next
	}
	fmt.Printf("双链表(正向): %s (长度: %d)\n", strings.Join(elements, " <-> "), l.Size)
}

// PrintBackward 反向打印
func (l *DoublyLinkedList) PrintBackward() {
	elements := make([]string, 0)
	current := l.Tail
	for current != nil {
		elements = append(elements, fmt.Sprintf("%d", current.Val))
		current = current.Prev
	}
	fmt.Printf("双链表(反向): %s (长度: %d)\n", strings.Join(elements, " <-> "), l.Size)
}

// ToSliceForward 转换为切片（正向）
func (l *DoublyLinkedList) ToSliceForward() []int {
	result := make([]int, 0, l.Size)
	current := l.Head
	for current != nil {
		result = append(result, current.Val)
		current = current.Next
	}
	return result
}

// ToSliceBackward 转换为切片（反向）
func (l *DoublyLinkedList) ToSliceBackward() []int {
	result := make([]int, 0, l.Size)
	current := l.Tail
	for current != nil {
		result = append(result, current.Val)
		current = current.Prev
	}
	return result
}

// ==================== 循环链表 ====================

// CircularLinkedList 循环链表
type CircularLinkedList struct {
	Head *ListNode
	Tail *ListNode
	Size int
}

// NewCircularLinkedList 创建循环链表
func NewCircularLinkedList() *CircularLinkedList {
	return &CircularLinkedList{
		Head: nil,
		Tail: nil,
		Size: 0,
	}
}

// Add 添加节点
func (l *CircularLinkedList) Add(val int) {
	node := &ListNode{Val: val}

	if l.Head == nil {
		l.Head = node
		l.Tail = node
		node.Next = node // 指向自己
	} else {
		l.Tail.Next = node
		node.Next = l.Head
		l.Tail = node
	}
	l.Size++
}

// Print 打印循环链表
func (l *CircularLinkedList) Print(maxDisplay int) {
	if l.Head == nil {
		fmt.Println("循环链表: 空")
		return
	}

	elements := make([]string, 0)
	current := l.Head
	count := 0

	for current != nil && count < maxDisplay {
		elements = append(elements, fmt.Sprintf("%d", current.Val))
		current = current.Next
		count++
		if current == l.Head {
			break
		}
	}

	if count < l.Size {
		fmt.Printf("循环链表: %s -> ... (循环, 总长度: %d)\n", strings.Join(elements, " -> "), l.Size)
	} else {
		fmt.Printf("循环链表: %s -> (回到头节点, 长度: %d)\n", strings.Join(elements, " -> "), l.Size)
	}
}

// ==================== 示例和测试 ====================

func main() {
	fmt.Println("==================== Go语言链表完整实现 ====================\n")

	// ==================== 单链表示例 ====================
	fmt.Println("1. 单链表基本操作")
	fmt.Println(strings.Repeat("-", 50))

	list := NewLinkedList()

	// 添加元素
	list.AddAtHead(1)
	list.AddAtTail(3)
	list.AddAtIndex(1, 2)
	list.Print()

	// 获取元素
	if val, ok := list.Get(1); ok {
		fmt.Printf("索引1的值: %d\n", val)
	}

	// 查找元素
	fmt.Printf("值2的位置: %d\n", list.Find(2))

	// 更新元素
	list.Update(1, 5)
	list.Print()

	// 删除元素
	list.DeleteAtIndex(1)
	list.Print()

	list.DeleteByValue(1)
	list.Print()

	// 反转链表
	fmt.Println("\n2. 单链表反转")
	list2 := NewLinkedList()
	for i := 1; i <= 5; i++ {
		list2.AddAtTail(i)
	}
	fmt.Print("反转前: ")
	list2.Print()
	list2.Reverse()
	fmt.Print("反转后: ")
	list2.Print()

	// 合并两个有序链表
	fmt.Println("\n3. 合并两个有序链表")
	l1 := NewLinkedList()
	l2 := NewLinkedList()

	for i := 1; i <= 5; i += 2 {
		l1.AddAtTail(i)
	}
	for i := 2; i <= 6; i += 2 {
		l2.AddAtTail(i)
	}

	fmt.Print("链表1: ")
	l1.Print()
	fmt.Print("链表2: ")
	l2.Print()

	merged := Merge(l1, l2)
	fmt.Print("合并后: ")
	merged.Print()

	// 删除倒数第n个节点
	fmt.Println("\n4. 删除倒数第n个节点")
	l3 := NewLinkedList()
	for i := 1; i <= 5; i++ {
		l3.AddAtTail(i)
	}
	fmt.Print("原始链表: ")
	l3.Print()

	l3.RemoveNthFromEnd(2)
	fmt.Print("删除倒数第2个节点后: ")
	l3.Print()

	// 获取中间节点
	fmt.Println("\n5. 获取中间节点")
	l4 := NewLinkedList()
	for i := 1; i <= 6; i++ {
		l4.AddAtTail(i)
	}
	fmt.Print("链表: ")
	l4.Print()
	middle := l4.GetMiddle()
	fmt.Printf("中间节点值: %d\n", middle.Val)

	// ==================== 双链表示例 ====================
	fmt.Println("\n6. 双链表操作")
	fmt.Println(strings.Repeat("-", 50))

	dlist := NewDoublyLinkedList()
	dlist.AddAtTail(1)
	dlist.AddAtTail(2)
	dlist.AddAtHead(0)
	dlist.AddAtTail(3)

	dlist.PrintForward()
	dlist.PrintBackward()

	dlist.DeleteByValue(2)
	fmt.Print("删除2后: ")
	dlist.PrintForward()

	// ==================== 循环链表示例 ====================
	fmt.Println("\n7. 循环链表操作")
	fmt.Println(strings.Repeat("-", 50))

	clist := NewCircularLinkedList()
	for i := 1; i <= 5; i++ {
		clist.Add(i)
	}
	clist.Print(10)

	// ==================== 约瑟夫问题 ====================
	fmt.Println("\n8. 约瑟夫问题（使用循环链表）")
	josephus(7, 3)

	// ==================== 性能测试 ====================
	fmt.Println("\n9. 性能测试")
	fmt.Println(strings.Repeat("-", 50))

	testList := NewLinkedList()
	n := 10000

	// 添加测试
	for i := 0; i < n; i++ {
		testList.AddAtTail(i)
	}
	fmt.Printf("添加 %d 个元素后，链表长度: %d\n", n, testList.GetSize())

	// 查找测试
	index := testList.Find(n - 1)
	fmt.Printf("查找最后一个元素的位置: %d\n", index)

	// 删除测试
	testList.DeleteAtIndex(n - 1)
	fmt.Printf("删除最后一个元素后，长度: %d\n", testList.GetSize())
}

// josephus 约瑟夫环问题
func josephus(n, k int) {
	if n <= 0 {
		return
	}

	// 创建循环链表
	clist := NewCircularLinkedList()
	for i := 1; i <= n; i++ {
		clist.Add(i)
	}

	fmt.Printf("约瑟夫环问题: %d个人，每次数到%d的人出局\n", n, k)

	current := clist.Head
	prev := clist.Tail

	result := make([]int, 0)

	for clist.Size > 0 {
		// 数k步
		for i := 1; i < k; i++ {
			prev = current
			current = current.Next
		}

		// 移除当前节点
		result = append(result, current.Val)
		prev.Next = current.Next

		if current == clist.Head {
			clist.Head = current.Next
		}
		if current == clist.Tail {
			clist.Tail = prev
		}

		current = current.Next
		clist.Size--
	}

	fmt.Printf("出局顺序: %v\n", result)
}
