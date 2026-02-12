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

package main

var a = b // 可以使用其后声明的变量的标识符 go1.24不行
var b = 123

func main() {
	// 下面两行中右边的标识符为预声明的标识符。
	const iota = iota // ok
	var true = true   // ok
	_ = true
}
