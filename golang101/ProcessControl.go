//package main
//
//import (
//	"fmt"
//	"math/rand"
//	"time"
//)
//
//func main() {
//	rand.Seed(time.Now().UnixNano()) // Go 1.20之前需要
//	switch n := rand.Intn(100); n % 9 {
//	case 0:
//		fmt.Println(n, "is a multiple of 9.")
//
//		// 和很多其它语言不一样，程序不会自动自动从
//		// 当前分支代码块跳到下一个分支代码块去执行。
//		// 所以，这里不需要一个break语句。
//	case 1, 2, 3:
//		fmt.Println(n, "mod 9 is 1, 2 or 3.")
//		break // 这里的break语句可有可无的，效果
//		// 是一样的。执行不会跳到下一个分支。
//	case 4, 5, 6:
//		fmt.Println(n, "mod 9 is 4, 5 or 6.")
//	// case 6, 7, 8:
//	// 上一行可能编译不过，因为6和上一个case中的
//	// 6重复了。是否能编译通过取决于具体编译器实现。
//	default:
//		fmt.Println(n, "mod 9 is 7 or 8.")
//	}
//}

/*
包含跳转标签的break和continue语句
如果一条break语句中包含一个跳转标签名，则此跳转标签必须刚好声明在一个包含此break语句的可跳出流程控制代码块之前，break语句将立即结束此跳出流程代码块的执行
如果一条continue语句中包含一个跳转标签名，则此跳转标签必须刚好声明在一个包含此continue语句的循环流程控制代码之前，continue语句将提前结束此循环流程控制代码块的
当前步的执行
*/
package main

import "fmt"

func FindSmallestPrimeLargerThan(n int) int {
Outer:
	for n++; ; n++ {
		for i := 2; ; i++ {
			switch {
			case i*i > n:
				break Outer
			case n%i == 0:
				continue Outer
			}
		}
	}
	return n
}

func main() {
	for i := 90; i < 100; i++ {
		n := FindSmallestPrimeLargerThan(i)
		fmt.Print("最小的比", i, "大的素数为", n)
		fmt.Println()
	}
}
