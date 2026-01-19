/*
对于标准编译器，字符串类型的内部结构声明如下：

	type _string struct {
		elements *byte // 引用着底层的字节
		len      int   // 字符串中的字节数
	}

从这个声明来看，我们可以将一个字符串的内部定义看作为一个字节序列。可以把一个字符串看作是一个元素类型为byte的（且元素不可修改的）切片。
byte是内置类型uint8的一个别名。
字符串值（和布尔以及各种数值类型的值）可以被用做常量。
Go支持两种风格的字符串字面量表示形式：双引号风格（解释型字面表示）和反引号风格（直白字面表示）。具体介绍请阅读前文。
字符串类型的零值为空字符串。一个空字符串在字面上可以用""或者“来表示。
我们可以用运算符+和+=来衔接字符串。
字符串类型都是可比较类型。同一个字符串类型的值可以用==和!=比较运算符来比较。
并且和整数/浮点数一样，同一个字符串类型的值也可以用>、<、>=和<=比较运算符来比较。
当比较两个字符串值的时候，它们的底层字节将逐一进行比较。如果一个字符串是另一个字符串的前缀，并且另一个字符串较长，则另一个字符串为两者中的较大者。
*/
package main

import "fmt"

func main() {
	const World = "world"
	var hello = "hello"
	var helloWorld = hello + "" + World
	helloWorld += "!"
	fmt.Println(helloWorld)
	fmt.Println(hello == "hello")
	fmt.Println(hello > helloWorld)
}
