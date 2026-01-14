//package main
//
//// 一些非简单语句：
//import "time"
//var a = 123
//const B = "Go"
//type Choice bool
//func f() int {
//	for a < 10 {
//		break
//	}
//
//	// 这是一个显式代码块。
//	{
//		// ...
//	}
//	return 567
//}
//
//// 一些简单语句的例子：
//c := make(chan bool) // 通道将在以后讲解
//a = 789
//a += 5
//a = f() // 这是一个纯赋值语句
//a++
//a--
//c <- true // 一个通道发送操作
//z := <-c  // 一个使用通道接收操作
//// 做为源值的变量短声明语句
//
//// 一些表达式的例子：
//123
//true
//B
//B + " language"
//a - 789
//a > 0 // 一个类型不确定布尔值
//f     // 一个类型为“func ()”的表达式
//
//// 下面这些即可以被视为简单语句，也可以被视为表达式。
//f() // 函数调用
//<-c // 通道接收操作
package main

import "fmt"

func main() {
	fmt.Println("Expression")
}
