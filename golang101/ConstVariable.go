/*
在常量声明中使用iota
iota是Go中预声明（内置）的一个特殊的具名常量。 iota被预声明为0，但是它的值在编译阶段并非恒定。
当此预声明的iota出现在一个常量声明中的时候，它的值在第n个常量描述中的值为n（从0开始）。 所以iota只对含有多个常量描述的常量声明有意义。
iota和常量描述自动补全相结合有的时候能够给Go编程带来很大便利。 比如，下面是一个使用了这两个特性的例子。 请阅读代码注释以了解清楚各个常量被绑定的值。
*/

//package main
//
//func main() {
//	const (
//		k = 3 // 在此处，iota == 0
//
//		m float32 = iota + .5 // m float32 = 1 + .5
//		n                     // n float32 = 2 + .5
//
//		p    = 9          // 在此处，iota == 3
//		q    = iota * 2   // q = 4 * 2
//		_                 // _ = 5 * 2
//		r                 // r = 6 * 2
//		s, t = iota, iota // s, t = 7, 7
//		u, v              // u, v = 8, 8
//		_, w              // _, w = 9, 9
//	)
//
//	const x = iota // x = 0 （iota == 0）
//	const (
//		y = iota // y = 0 （iota == 0）
//		z        // z = 1
//	)
//
//	println(m)             // +1.500000e+000
//	println(n)             // +2.500000e+000
//	println(q, r)          // 8 12
//	println(s, t, u, v, w) // 7 7 8 8 9
//	println(x, y, z)       // 0 0 1
//}

//package main
//
//func main() {
//	// 变量lang和year都为新声明的变量。
//	lang, year := "Go language", 2007
//
//	// 这里，只有变量createdBy是新声明的变量。
//	// 变量year已经在上面声明过了，所以这里仅仅
//	// 改变了它的值，或者说它被重新声明了。
//	year, createdBy := 2009, "Google Research"
//
//	// 这是一个纯赋值语句。
//	lang, year = "Go", 2012
//
//	print(lang, "由", createdBy, "发明")
//	print("并发布于", year, "年。")
//	println()
//}

/*
在Go中，我们可以使用一对大括号来显式形成一个（局部）代码块。一个代码块可以内嵌另一个代码块。 最外层的代码块称为包级代码块。
一个声明在一个内层代码块中的常量或者变量将遮挡另一个外层代码块中声明的同名变量或者常量。
比如，下面的代码中声明了3个名为x的变量。 内层的x将遮挡外层的x， 从而外层的x在内层的x声明之后在内层中将不可见。
*/
//package main
//
//const y = 70
//
//var x int = 123 // 包级变量
//
//func main() {
//	// 此x变量遮挡了包级变量x。
//	var x = true
//
//	// 一个内嵌代码块。
//	{
//		x, y := x, y-10 // 这里，左边的x和y均为新声明
//		// 的变量。右边的x为外层声明的
//		// bool变量。右边的y为包级变量。
//
//		// 在此内层代码块中，从此开始，
//		// 刚声明的x和y将遮挡外层声明x和y。
//
//		x, z := !x, y/10 // z是一个新声明的变量。
//		// x和y是上一句中声明的变量。
//		println(x, y, z) // false 60 6
//	}
//	println(x) // true
//	println(y) // 70 （包级变量y从未修改）
//	/*
//		println(z) // error: z未定义。
//		           // z的作用域仅限于上面的最内层代码块。
//	*/
//}

/*
一个类型不确定常量所表示的值可以溢出其默认类型
*/

//package main
//
//const n = 1 << 64
//const r = 'a' + 0x7FFFFFFF
//const x = 2e+308
//
//func main() {
//	_ = n >> 2
//	_ = r - 0x7FFFFFFF
//	_ = x / 2
//}

/*
每个常量标识符可以看做是增强型的C语言中的#define宏，在编译阶段，所有的标识符将被他们各自绑定的字面量所替代
若一个运算中所有的运算数都为常量，运算结果也为常量
*/
package main

const X = 3
const Y = X + X

var a = X

func main() {
	b := Y
	println(a, b, X, Y)
}
