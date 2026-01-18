/*
Go语言中，一个结构类型的尺寸为它所有字段的尺寸之和加上一些填充字节的数目，通常编译器会在一个结构体值的两个相邻字段之间填充一些字节来保证一些字段的地址
总是某个整数的倍数。一个零字段结构体的尺寸为零。每个结构体字段在它的声明中可以被指定一个标签，字段标签可以是任意字符串，默认为空字符串。
在实际中，字段标签表示成用空格分隔的键值对形式，标签尽量用直白字面形式（“）键值对中使用解释型字面形式（""）
通常不推荐把字段标签当成字段注释使用，同时Go结构体不支持字段联合。结构体类型中字段标签和字段的声明顺序对结构体身份识别很重要，两个无名结构体类型的各个对应字段声明
只有在名称、类型、标签以及内嵌性都等同的情况下才相同。
*/
// 对于类型S的一个值v，我们可以用v.x和v.y来表示它的字段。 v.x（或v.y）这种形式称为一个选择器（selector）。
// 其中的v称为此选择器的属主。 今后，我们称一个选择器中的句点.为属性选择操作符。
//package main
//
//import "fmt"
//
//type Book struct {
//	title, author string
//	pages         int
//}
//
//func main() {
//	book := Book{"go", "l", 256}
//	fmt.Println(book)
//	// 使用带字段名的组合字面量来表示结构体的值
//	book = Book{author: "l", pages: 256, title: "go"}
//	book = Book{}
//	book = Book{author: "l"}
//	var book2 Book
//	book2.title = "go"
//	book2.pages = 300
//	fmt.Println(book2.pages)
//}

/*
结构体字段的可寻址性:如果一个结构体值是可寻址的，则它的字段也是可寻址的；
反之，一个不可寻址的结构体值的字段也是不可寻址的。 不可寻址的字段的值是不可更改的。所有的组合字面量都是不可寻址的。
可寻址的：赋值给变量的结构体是可寻址的，变量/切片中的元素/指针解引用后都是可寻址的
不可寻址的：临时变量、map中的元素
*/
//package main
//
//import "fmt"
//
//func main() {
//	type Book struct {
//		Pages int
//	}
//	var book = Book{}
//	p := &book.Pages //选择器中的属性选择操作符.的优先级比取地址操作符&的优先级要高。
//	*p = 123
//	fmt.Println(book)
//}

// 一般只有可被寻址的值才能被取地址，
// 但是Go因为语法糖使得不可被寻址的组合字面量可以被取地址
//package main
//
//func main() {
//	type Book struct {
//		Pages int
//	}
//	p := &Book{100}
//	p.Pages = 200
//}

// 此外，在字段选择器中，属主结构体值可以是指针，将被隐式解引用
//package main
//
//func main() {
//	type Book struct {
//		pages int
//	}
//	book1 := &Book{100}
//	book2 := new(Book)
//	book2.pages = book1.pages
//	(*book2).pages = (*book1).pages
//}

package main

import (
	"fmt"
	"reflect"
)

type S0 struct {
	y int "foo"
	x bool
}

type S1 = struct { // S1是一个无名类型
	x int "foo"
	y bool
}

type S2 = struct { // S2也是一个无名类型
	x int "bar"
	y bool
}

type S3 S2 // S3是一个定义类型（因而具名）。
type S4 S3 // S4是一个定义类型（因而具名）。
// 如果不考虑字段标签，S3（S4）和S1的底层类型一样。
// 如果考虑字段标签，S3（S4）和S1的底层类型不一样。

func main() {
	// 初始化变量
	var v0 S0
	var v1 S1
	var v2 S2
	var v3 S3
	var v4 S4

	// 打印函数
	printTypeInfo := func(name string, v interface{}) {
		t := reflect.TypeOf(v)
		fmt.Printf("变量: %s\n", name)
		fmt.Printf("  fmt.Printf(\"%%T\"): %T\n", v)
		fmt.Printf("  reflect Name():     %s\n", t.Name())   // 类型名称
		fmt.Printf("  reflect String():   %s\n", t.String()) // 完整类型描述
		fmt.Println("--------------------------------------------------")
	}

	printTypeInfo("v0 (S0)", v0)
	printTypeInfo("v1 (S1)", v1)
	printTypeInfo("v2 (S2)", v2)
	printTypeInfo("v3 (S3)", v3)
	printTypeInfo("v4 (S4)", v4)
}

/*
因为 `S0` 有了自己的名字，Go 认为这个名字比它底层的实现细节（struct）更重要。
让我们深入剖析一下其中的逻辑：
1. 具名类型（Named Type） vs 未命名类型（Unnamed Type）
在 Go 语言中，类型分为两大类：
具名类型（Named Type）：通过 `type Name Type` 定义的类型。例如 `type S0 struct {...}`。
未命名类型（Unnamed Type）：直接写出的类型字面量。例如 `struct { x int }`、`[]int`、`*int`、`map[string]int`。

2. 为什么 `main.S0` 不显示 `struct`？
当你定义 `type S0 struct { ... }` 时，你实际上是在告诉编译器：
> "请创造一个新的类型，它的名字叫 `S0`。虽然它的底层实现是一个结构体，但在类型系统中，**它就是 S0**。"
对于 `%T` 或 `reflect.Type.String()` 来说，它们的任务是打印出该类型的**唯一标识符**。
对于 S0：它的唯一标识符就是它的名字（包含包路径），即 `main.S0`。这就足够区分它和其他类型了。如果打印成 `struct main.S0` 或者是 `struct { ... }`，反而混淆了视听，或者丢失了它作为“专用类型”的身份。
对于 S1：它是 `type S1 = struct { ... }`。因为使用了 `=`（别名），`S1` 并没有创造新类型，它只是那个匿名结构体的“绰号”。那个匿名结构体没有名字，所以它的唯一标识符只能是它的**完整结构描述**，即 `struct { x int "foo"; y bool }`。
3. 类比理解
这就好比**品牌**和**配料表**的区别：
**S0 (具名类型) 就像 "可口可乐"**。
    当你问服务员这是什么饮料时，服务员会说：“这是**可口可乐**（main.S0）”。他不会念配料表说：“这是水、糖、碳酸、焦糖色...的混合物”。因为“可口可乐”这个名字已经包含了所有信息，并且代表了特定的身份。
**S1 (匿名结构体) 就像 "路边摊自制饮料"**。
    它没有注册商标。当你问这是什么时，老板只能描述它的成分：“这是**水加糖加柠檬**（struct { ... }）”。因为它没有专属的名字，只能用成分来描述它。
4. 如何强行看到它是 struct？
虽然 `%T` 隐藏了 `struct` 这个词，但 `S0` 的本质依然是 struct。如果你想在代码中确认它是 struct，或者查看它的底层结构，可以使用 `reflect` 包的 `Kind()` 方法：

package main

import (
	"fmt"
	"reflect"
)

type S0 struct {
	y int "foo"
	x bool
}

func main() {
	var v0 S0
	t := reflect.TypeOf(v0)

	// 1. 打印类型名 (Name)
	fmt.Println("Type Name:", t.String()) // 输出: main.S0

	// 2. 打印底层种类 (Kind) -> 这里会告诉你它是 struct
	fmt.Println("Underlying Kind:", t.Kind()) // 输出: struct
}
*/

/*
匿名结构体类型允许出现在结构体字段声明中。匿名结构体类型也允许出现在组合字面量中。

一个例子：
var aBook = struct {
	author struct { // 此字段的类型为一个匿名结构体类型
		firstName, lastName string
		gender              bool
	}
	title string
	pages int
}{
	author: struct {
		firstName, lastName string
		gender              bool
	}{
		firstName: "Mark",
		lastName: "Twain",
	}, // 此组合字面量中的类型为一个匿名结构体类型
	title: "The Million Pound Note",
	pages: 96,
}
通常来说，为了代码可读性，最好少使用匿名结构体类型。
*/
