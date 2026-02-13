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

package main

import "fmt"

var f = func(b bool) {
	fmt.Print("Goat")
}

func main() {
	var f = func(b bool) {
		fmt.Print("Sheep")
		if b {
			fmt.Print(" ")
			f(!b) // 此f乃包级变量f也。
		}
	}
	f(true) // 此f为刚声明的局部变量f。
}
