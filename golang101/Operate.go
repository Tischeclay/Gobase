/*
清位运算符&^
*/
//package main
//
//import "fmt"
//
//const (
//	Readable   = 1 << 0 // 0001
//	Writable   = 1 << 1 // 0010
//	Executable = 1 << 2 // 0100
//)
//
//func main() {
//	// 初始权限：可读、可写、可执行
//	perm := Readable | Writable | Executable
//	fmt.Printf("初始权限: %04b\n", perm) // 0111
//
//	// 移除 "写" 权限
//	// 逻辑：Writable (0010) 的第2位是1，所以 perm 的第2位会被强制清零
//	perm = perm &^ Writable
//	fmt.Printf("移除写后: %04b\n", perm) // 0101
//
//	// 再次尝试移除 "写" 权限 (即使本来就没有，也不会出错，保持原样)
//	perm = perm &^ Writable
//	fmt.Printf("再次移除: %04b\n", perm) // 0101
//}
package main

func main() {
	var (
		a, b float32 = 12.0, 3.14
		c, d int16   = 15, -6
		e    uint8   = 7
	)

	// 这些行编译没问题。
	_ = 12 + 'A' // 两个类型不确定操作数（都为数值类型）
	_ = 12 - a   // 12将被当做a的类型（float32）使用。
	_ = a * b    // 两个同类型的类型确定操作数。
	_ = c % d
	_, _ = c+int16(e), uint8(c)+e
	_, _, _, _ = a/b, c/d, -100/-9, 1.23/1.2
	_, _, _, _ = c|d, c&d, c^d, c&^d
	_, _, _, _ = d<<e, 123>>e, e>>3, 0xF<<0
	_, _, _, _ = -b, +c, ^e, ^-1

	// 这些行编译将失败。
	//_ = a % b   // error: a和b都不是整数
	//_ = a | b   // error: a和b都不是整数
	//_ = c + e   // error: c和e的类型不匹配
	//_ = b >> 5  // error: b不是一个整数
	//_ = c >> -5 // error: -5不是一个无符号整数

	_ = e << uint(c) // 编译没问题
	_ = e << c       // 从Go 1.13开始，此行才能编译过
	_ = e << -c      // 从Go 1.13开始，此行才能编译过。
	// 将在运行时刻造成恐慌。
	//_ = e << -1      // error: 右操作数不能为负（常数）
}
