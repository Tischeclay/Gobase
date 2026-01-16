// 加入分号后被ide自动忽略并清除了
//package main
//
//import "fmt"
//
//func main() {
//	var (
//		i   int
//		sum int
//	)
//	for i < 6 {
//		sum += i
//		i++
//	}
//	fmt.Println(sum)
//}

/*在Go代码中，以下断行是没问题的（不影响程序行为的）：
在除了break、continue和return这几个跳转关键字之外的任何关键字之后断行，或者在不跟随标签的break和continue关键字以及不跟随返回值的return关键字之后断行；
在（显式输入的或者隐式被编译器插入的）分号;之后断行；
在不会导致新的隐式分号被编译器插入的情况下断行。*/

// 以下代码块是合法的,编译器将在这些行的行尾自动插入一个分号：第9行、第10行、第15行和第20行。
package main

import "fmt"

func alwaysFalse() bool { return false }

func main() {
	for i := 0; i < 6; i++ {
		// 使用i ...
	}

	if x := alwaysFalse(); !x {
		// ...
	}

	switch alwaysFalse(); {
	case true:
		fmt.Println("true")
	case false:
		fmt.Println("false")
	}
}
