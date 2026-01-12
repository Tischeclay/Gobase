//package main
//
//func SquaresOfSumAndDiff(a int64, b int64) (int64, int64) {
//	return (a + b) * (a + b), (a - b) * (a - b)
//}
//
//func CompareLower4bits(m, n uint32) (r bool) {
//	r = m&0xF > n&0xf
//	return
//}
//
//// 使用一个函数调用的返回结果来初始化一个包级变量。
//var v = VersionString()
//
//func main() {
//	println(v) // v1.0
//	x, y := SquaresOfSumAndDiff(3, 6)
//	println(x, y) // 81 9
//	b := CompareLower4bits(uint32(x), uint32(y))
//	println(b) // false
//	// "Go"的类型被推断为string；1的类型被推断为int32。
//	doNothing("Go", 1)
//}
//
//func VersionString() string {
//	return "v1.0"
//}
//
//func doNothing(string, int32) {
//}

package main

func main() {
	// 这个匿名函数没有输入参数，但是返回两个结果
	x, y := func() (int, int) {
		println("This function has no parameters")
		return 3, 4
	}() // 一对小括号表示立即调用此函数，不需要传递实参

	// 以下这些匿名函数没有返回结果
	func(a, b int) {
		println("a * a + b * b =", a*a+b*b)
	}(x, y) // 立即调用该匿名函数并传递两个实参

	func(x int) {
		println("x * x + y * y = ", x*x+y*y)
	}(y) // 将实参y传递给形参x

	func() {
		println("x * x + y * y = ", x*x+y*y)
	}() // 不需要传递实参
}
