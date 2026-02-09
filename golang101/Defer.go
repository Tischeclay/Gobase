//package main
//
//import "fmt"
//
//func main() {
//	s := []string{"a", "b", "c", "d"}
//	defer fmt.Println(s) // [a x y d]
//	// defer append(s[:1], "x", "y") // 编译错误
//	defer func() {
//		_ = append(s[:1], "x", "y")
//	}()
//}

//package main
//
//import "fmt"
//
//func main() {
//	defer fmt.Println("此行可以被执行到")
//	var f func() // f == nil
//	defer f()    // 将产生一个恐慌
//	fmt.Println("此行可以被执行到")
//	f = func() {} // 此行不会阻止恐慌产生
//}

package main

type T int

func (t T) M(n int) T {
	print(n)
	return t
}

func main() {
	var t T
	// t.M(1)是方法调用M(2)的属主实参，因此它
	// 将在M(2)调用被推入延迟调用栈时被估值。
	defer t.M(1).M(2)
	t.M(3).M(4)
}
