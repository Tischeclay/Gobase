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

package main

import "fmt"

func main() {
	defer fmt.Println("此行可以被执行到")
	var f func() // f == nil
	defer f()    // 将产生一个恐慌
	fmt.Println("此行可以被执行到")
	f = func() {} // 此行不会阻止恐慌产生
}
