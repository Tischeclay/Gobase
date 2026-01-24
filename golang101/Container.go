/*
在严格意义上，Go中有三种一等公民容器类型：数组、切片和映射。
有时候，我们可以认为字符串类型和通道类型也属于容器类型。 但是，此篇文章只谈及数组、切片和映射类型。
存储在一个容器中的每个元素值都关联着一个键值（key）。每个元素可以通过它的键值而被访问到。 一个映射类型的键值类型必须为一个可比较类型。 数组和切片类型的键值类型均为内置类型int。
一个数组或切片的一个元素对应的键值总是一个非负整数下标，此非负整数表示该元素在该数组或切片所有元素中的顺序位置。此非负整数下标亦常称为一个元素索引。
在官方标准编译器和运行时中，映射是使用哈希表算法来实现的。所以一个映射中的所有元素也均存放在一块连续的内存中，但是映射中的元素并不一定紧挨着存放。
另外一种常用的映射实现算法是二叉树算法。无论使用何种算法，一个映射中的所有元素的键值也存放在此映射值（的间接部分）中
无名容器类型的字面表示形式如下：
数组类型：[N]T
切片类型：[]T
映射类型：map[K]T
其中，
T可为任意类型。它表示一个容器类型的元素类型。某个特定容器类型的值中只能存储此容器类型的元素类型的值。
N必须为一个非负整数常量。它指定了一个数组类型的长度，或者说它指定了此数组类型的任何一个值中存储了多少个元素。 一个数组类型的长度是此数组类型的一部分。
比如[5]int和[8]int是两个不同的类型。K必须为一个可比较类型。它指定了一个映射类型的键值类型。
*/
//const Size = 32
//
//type Person struct {
//	name string
//	age int
//}
//
//[5]string
//[Size]int
//[16][]byte   // 元素类型为切片类型
//[100]Person  // 元素类型为一个结构体类型
//
//[]bool
//[]int64
//[]map[int]bool
//[]*int
//
////
//map[string]int
//map[int]bool
//map[int16][6]string // 元素类型为一个数组类型
//map[bool][]string   // 元素类型为一个切片类型
//map[struct{x int}]*int8  // 元素类型为一个指针类型，键值类型为一个结构体类型

/*
映射组合字面量中大括号中的每一项称为一个键值元素对（key-value pair），或者称为一个条目（entry）
// 下面这些切片字面量都是等价的。
[]string{"break", "continue", "fallthrough"}
[]string{0: "break", 1: "continue", 2: "fallthrough"}
[]string{2: "fallthrough", 1: "continue", 0: "break"}
[]string{2: "fallthrough", 0: "break", "continue"}
// 下面这些数组字面量都是等价的。
[4]bool{false, true, true, false}
[4]bool{0: false, 1: true, 2: true, 3: false}
[4]bool{1: true, true}
[4]bool{2: true, 1: true}
[...]bool{false, true, true, false}
[...]bool{3: false, 1: true, true}
其中...表示让编译器推断出相应数组值的类型的长度
如果一个索引下标出现，它的类型不必是数组和切片类型的键值类型int，但它必须是一个可以表示为int值的非负常量；
如果它是一个类型确定值，则它的类型必须为一个内置整数类型。
在一个数组或切片组合字面量中，如果一个元素的索引下标缺失，则编译器认为它的索引下标为出现在它之前的元素的索引下标加一。
如果出现的第一个元素的索引下标缺失，则它的索引下标被认为是0。
映射组合字面量中元素对应的键值不可缺失，并且它们可以为非常量。
和结构体类似，一个数组类型A的零值可以表示为A{}。 比如，数组类型[100]int的零值可以表示为[100]int{}。 一个数组零值中的所有元素均为对应数组元素类型的零值。
和指针一样，所有切片和映射类型的零值均用预声明的标识符nil来表示。
顺便说一句，除了刚提到的三种类型，以后将介绍的函数、通道和接口类型的零值也用预声明的标识符nil来表示。
在运行时刻，即使一个数组变量在声明的时候未指定初始值，它的元素所占的内存空间也已经被开辟出来。 但是一个nil切片或者映射值的元素的内存空间尚未被开辟出来。
注意：[]T{}表示类型[]T的一个空切片值，它和[]T(nil)是不等价的。 同样，map[K]T{}和map[K]T(nil)也是不等价的。
*/

// 容器字面量是不可寻址的但可以被取地址
//package main
//
//import "fmt"
//
//func main() {
//	pm := &map[string]int{"C": 1972, "Go": 2009}
//	ps := &[]string{"break", "continue"}
//	pa := &[...]bool{false, true, true, false}
//	fmt.Printf("%T\n", pm) // *map[string]int
//	fmt.Printf("%T\n", ps) // *[]string
//	fmt.Printf("%T\n", pa) // *[4]bool
//}

//在某些情形下，内嵌在其它组合字面量中的组合字面量可以简化为{...}（即类型部分被省略掉了）。 内嵌组合字面量前的取地址操作符&有时也可以被省略。

/*
	尽管两个映射值和切片值是不能比较的，但是一个映射值或者切片值可以和预声明的nil标识符进行比较以检查此映射值或者切片值是否为一个零值。

大多数数组类型都是可比较类型，除了元素类型为不可比较类型的数组类型。当比较两个数组值时，它们的对应元素将按照逐一被比较（可以认为按照下标顺序比较）。
这两个数组只有在它们的对应元素都相等的情况下才相等；当一对元素被发现不相等的或者在比较中产生恐慌的时候，对数组的比较将提前结束。
*/
//package main
//
//import "fmt"
//
//func main() {
//	var a [16]byte
//	var s []int
//	var m map[string]int
//
//	fmt.Println(a == a)                  // true
//	fmt.Println(m == nil)                // true
//	fmt.Println(s == nil)                // true
//	fmt.Println(nil == map[string]int{}) // false
//	fmt.Println(nil == []int{})          // false
//}

/*
除了上面已提到的容器长度属性（此容器中含有有多少个元素），每个容器值还有一个容量属性。 一个数组值的容量总是和它的长度相等；
一个非零映射值的容量可以被认为是无限大的。切片值的容量的含义将在后续章节介绍。 一个切片值的容量总是不小于此切片值的长度。
在编程中，只有切片值的容量有实际意义。
我们可以调用内置函数len来获取一个容器值的长度，或者调用内置函数cap来获取一个容器值的容量。
这两个函数都返回一个int类型确定结果值或者一个默认类型为int的类型不确定结果，具体取决于传递给它们的实参是否为常量表达式。
因为非零映射值的容量是无限大，所以cap并不适用于映射值。
*/
//package main
//
//import "fmt"
//
//func main() {
//	var a [5]int
//	fmt.Println(len(a), cap(a)) // 5 5
//	var s []int
//	fmt.Println(len(s), cap(s)) // 0 0
//	s, s2 := []int{2, 3, 5}, []bool{}
//	fmt.Println(len(s), cap(s), len(s2), cap(s2)) // 3 3 0 0
//	var m map[int]bool
//	fmt.Println(len(m)) // 0
//	m, m2 := map[int]bool{1: true, 0: false}, map[int]int{}
//	fmt.Println(len(m), len(m2)) // 2 0
//}

//package main
//
//import "fmt"
//
//func main() {
//	a := []int{-1, 0, 1}
//	s := []bool{true, false}
//	m := map[string]int{"abc": 123, "xyz": 789}
//	fmt.Println(a[2], s[1], m["abc"])
//	a[2], s[1], m["abc"] = 999, true, 567
//	fmt.Println(a[2], s[1], m["abc"])
//
//	n, present := m["hello"]
//	fmt.Println(n, present, m["hello"])
//	n, present = m["abc"]
//	fmt.Println(n, present, m["abc"])
//	//m["hello"] = 555
//}

// 当一个映射赋值语句执行完毕之后，目标映射值和源映射值将共享底层的元素。 向其中一个映射中添加（或从中删除）元素将体现在另一个映射中。
// 和映射一样，当一个切片赋值给另一个切片后，它们将共享底层的元素。它们的长度和容量也相等。 但是和映射不同，如果以后其中一个切片改变了长度或者容量，此变化不会体现到另一个切片中。
// 当一个数组被赋值给另一个数组，所有的元素都将被从源数组复制到目标数组。赋值完成之后，这两个数组不共享任何元素。
//package main
//
//import "fmt"
//
//func main() {
//	m0 := map[int]int{0: 7, 1: 8, 2: 9}
//	m1 := m0
//	m1[0] = 2
//	fmt.Println(m0, m1) // map[0:2 1:8 2:9] map[0:2 1:8 2:9]
//
//	s0 := []int{7, 8, 9}
//	s1 := s0
//	s1[0] = 2
//	fmt.Println(s0, s1) // [2 8 9] [2 8 9]
//
//	a0 := [...]int{7, 8, 9}
//	a1 := a0
//	a1[0] = 2
//	fmt.Println(a0, a1) // [7 8 9] [2 8 9]
//}

// 向一个映射添加和从一个映射删除条目
//package main
//
//import "fmt"
//
//func main() {
//	m := map[string]int{"Go": 2007}
//	m["C"] = 1972
//	m["Java"] = 1995
//	fmt.Println(m)
//	m["Go"] = 2009
//	delete(m, "Java")
//	fmt.Println(m)
//}

// 注意，在Go 1.12之前，映射打印结果中的条目顺序并不固定，两次打印结果可能并不相同。
// 一个数组中的元素个数总是恒定的，我们无法向其中添加元素，也无法从其中删除元素。但是可寻址的数组值中的元素是可以被修改的。
// 我们可以通过调用内置append函数，以一个切片为基础，来添加不定数量的元素并返回一个新的切片。
// 此新的结果切片包含着基础切片中所有的元素和所有被添加的元素。 注意，基础切片并未被此append函数调用所修改。
// 当然，如果我们愿意（事实上在实践中常常如此），我们可以将结果切片赋值给基础切片以修改基础切片。
// Go中并未提供一个内置方式来从一个切片中删除一个元素,使用append函数和后面将要介绍的子切片语法一起来实现元素删除操作。
//package main
//
//import "fmt"
//
//func main() {
//	s0 := []int{2, 3, 5}
//	fmt.Println(s0, cap(s0)) // [2 3 5] 3
//	// 内置append函数是一个变长参数函数（下下篇文章中介绍）。 它有两个参数，其中第二个参数（形参）为一个变长参数。
//	s1 := append(s0, 7)      // 添加一个元素
//	fmt.Println(s1, cap(s1)) // [2 3 5 7] 6
//	s2 := append(s1, 11, 13) // 添加两个元素
//	fmt.Println(s2, cap(s2)) // [2 3 5 7 11 13] 6
//	s3 := append(s0)         // <=> s3 := s0
//	fmt.Println(s3, cap(s3)) // [2 3 5] 3
//	s4 := append(s0, s0...)  // 以s0为基础添加s0中所有的元素
//	fmt.Println(s4, cap(s4)) // [2 3 5 2 3 5] 6
//
//	s0[0], s1[0] = 99, 789
//	fmt.Println(s2[0], s3[0], s4[0]) // 789 99 2
//}

// 请注意，当一个append函数调用需要为结果切片开辟内存时，结果切片的容量取决于具体编译器实现。
// 在这种情况下，对于官方标准编译器，如果基础切片的容量较小，则结果切片的容量至少为基础切片的两倍。
// 这样做的目的是使结果切片有足够多的冗余元素槽位，以防止此结果切片被用做后续其它append函数调用的基础切片时再次开辟内存。
// 上面提到了，在实际编程中，我们常常将append函数调用的结果赋值给基础切片。
//package main
//
//import "fmt"
//
//func main() {
//	var s = append([]string(nil), "array", "slice")
//	fmt.Println(s)      // [array slice]
//	fmt.Println(cap(s)) // 2
//	s = append(s, "map")
//	fmt.Println(s)      // [array slice map]
//	fmt.Println(cap(s)) // 4
//	s = append(s, "channel")
//	fmt.Println(s)      // [array slice map channel]
//	fmt.Println(cap(s)) // 4
//}

// 第一个函数调用创建了一个长度为length并且容量为capacity的切片。 第二个函数调用创建了一个长度为length并且容量也为length的切片。
// 使用make函数创建的切片中的所有元素值均被初始化为（结果切片的元素类型的）零值。
// 下面是一个展示了如何使用make函数来创建映射和切片的例子：
//package main
//
//import "fmt"
//
//func main() {
//	// 创建映射。
//	fmt.Println(make(map[string]int)) // map[]
//	m := make(map[string]int, 3)
//	fmt.Println(m, len(m)) // map[] 0
//	m["C"] = 1972
//	m["Go"] = 2009
//	fmt.Println(m, len(m)) // map[C:1972 Go:2009] 2
//
//	// 创建切片。
//	s := make([]int, 3, 5)
//	fmt.Println(s, len(s), cap(s)) // [0 0 0] 3 5
//	s = make([]int, 2)
//	fmt.Println(s, len(s), cap(s)) // [0 0] 2 2
//}

// 注意映射和切片可以使用内置make函数来创建映射和切片

/*
在前面的指针一文中，我们已经了解到内置new函数可以用来为一个任何类型的值开辟内存并返回一个存储有此值的地址的指针。 用new函数开辟出来的值均为零值。因为这个原因，new函数对于创建映射和切片值来说没有任何价值。
使用new函数来用来创建数组值并非是完全没有意义的，但是在实践中很少这么做，因为使用组合字面量来创建数组值更为方便。

*/
// 使用new函数创建容器值
package main

import "fmt"

func main() {
	m := *new(map[string]int)
	fmt.Println(m == nil)
	s := *new([]int)
	fmt.Println(s == nil)
	a := *new([5]bool)
	fmt.Println(a == [5]bool{})
}
