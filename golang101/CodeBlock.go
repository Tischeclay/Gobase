//package main
//
//func main() {
//	// var v int = v   // error: v未定义
//	// const C int = C // error: C未定义
//	/*
//		type T = struct {
//			*T    // error: 不可循环引用
//			x []T // error: 不可循环引用
//		}
//	*/
//
//	// 下面所有的类型定义声明都是合法的。
//	type T struct {
//		*T
//		x []T
//	}
//	type A [5]*A
//	type S []S
//	type M map[int]M
//	type F func(F) F
//	type Ch chan Ch
//	type P *P
//
//	// ...
//	var p P
//	p = &p
//	p = ***********************p
//	***********************p = p
//
//	var s = make(S, 3)
//	s[0] = s
//	s = s[0][0][0][0][0][0][0][0]
//
//	var m = M{}
//	m[1] = m
//	m = m[1][1][1][1][1][1][1][1]
//}

//package main
//
//var a = b // 可以使用其后声明的变量的标识符 go1.24不行
//var b = 123
//
//func main() {
//	// 下面两行中右边的标识符为预声明的标识符。
//	const iota = iota // ok
//	var true = true   // ok
//	_ = true
//}

//package main
//
//import "fmt"
//
//var p0, p1, p2, p3, p4, p5 *int
//var x = 9999 // x#0
//
//func main() {
//	p0 = &x
//	var x = 888 // x#1
//	p1 = &x
//	for x := 70; x < 77; x++ { // x#2
//		p2 = &x
//		x := x - 70 //  // x#3
//		p3 = &x
//		if x := x - 3; x > 0 { // x#4
//			p4 = &x
//			x := -x // x#5
//			p5 = &x
//		}
//	}
//
//	// 9999 888 77 6 3 -3
//	fmt.Println(*p0, *p1, *p2, *p3, *p4, *p5)
//}

//package main
//
//import "fmt"
//
//var f = func(b bool) {
//	fmt.Print("Goat")
//}
//
//func main() {
//	var f = func(b bool) {
//		fmt.Print("Sheep")
//		if b {
//			fmt.Print(" ")
//			f(!b) // 此f乃包级变量f也。
//		}
//	}
//	f(true) // 此f为刚声明的局部变量f。
//}

//package main
//
//import "fmt"
//import "strconv"
//
//func parseInt(s string) (int, error) {
//	n, err := strconv.Atoi(s)
//	if err != nil {
//		// 一些新手Go程序员会认为下一行中声明
//		// 的err变量已经在外层声明过了。然而其
//		// 实下一行中的b和err都是新声明的变量。
//		// 此新声明的err遮挡了外层声明的err。
//		b, err := strconv.ParseBool(s)
//		if err != nil {
//			return 0, err
//		}
//
//		// 如果代码运行到这里，一些新手Go程序员
//		// 期望着内层的nil err将被返回。但是其实
//		// 返回是外层的非nil err。因为内层的err
//		// 的作用域到外层if代码块结尾就结束了。
//		if b {
//			n = 1
//		}
//	}
//	return n, err
//}
//
//func main() {
//	fmt.Println(parseInt("TRUE"))
//}

package main

import (
	"fmt"
)

const len = 3     // 遮挡了内置函数len
var true = 0      // 遮挡了内置常量true
type nil struct{} // 遮挡了内置变量nil
func int()        {} // 遮挡了内置类型int

func main() {
	fmt.Println("a weird program")
	var output = fmt.Println

	var fmt = [len]nil{{}, {}, {}} // 遮挡了包引入fmt
	// var n = len(fmt) // error: len是一个常量
	var n = cap(fmt) // 我们只好使用内置cap函数

	// for关键字跟随着一个隐式代码块和一个显式代码块。
	// 变量短声明中的true遮挡了全局变量true。
	for true := 0; true < n; true++ {
		// 下面声明的false遮挡了内置常量false。
		var false = fmt[true]
		// 下面声明的true遮挡了循环变量true。
		var true = true + 1
		// 下一行编译不通过，因为fmt是一个数组。
		// fmt.Println(true, false)
		output(true, false)
	}
}
