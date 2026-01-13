/*
内置函数print和println提供了和fmt标准库包中的对应函数相似的功能。 内置函数可以不用引入任何代码包而直接使用。
注意：print和println这两个内置函数不推荐使用在生产环境，因为它们不保证一定会出现在以后的Go版本中。
*/
//package main
//
//import "fmt"
//
//func main() {
//	a, b := 123, "Go"
//	fmt.Printf("a == %v == 0x%x, b = %s\n", a, a, b)
//	fmt.Printf("type of a : %T, type of b : %T\n", a, b)
//	fmt.Printf("1%% 50%% 99%%\n")
//}

/*
关于代码包目录、代码包引入路径和代码包依赖关系
一个代码包可以由若干个Go源文件组成，一个代码包的源文件必须处于同一个目录下，一个目录下所有源文件（不包含子目录）必须处于同一代码包中
也就是说这些源文件开头的package pkgname语句必须一致。一个代码包对应一个目录，反之亦然，代码包目录下子目录对应的是另外的独立的代码包
在Go中，一个引入路径包含internal目录名的代码包被视为一个特殊的代码包，只能被internal目录的直接父目录（和此父目录的子目录）中的代码包所引入。
Go不支持循环依赖
*/
/*
init函数：在一个代码包、一个源文件中可以声明若干名为init的函数，这些函数必须不带任何输入参数和返回结果
不能声明名为Init的包级变量、常量或者类型
在程序运行时刻，在进入main入口函数之前，每个init函数在此包加载的时候被串行执行一遍
*/
/*
在加载一个代码包的过程中，所有的声明在此包中的init函数将被串行调用且仅调用执行一次，一个代码包中声明的init函数的调用肯定晚于此代码包所依赖的代码包中声明的init函数
同一个源文件中声明的init函数将按从上到下的顺序被执行调用，对于声明在同一个包的两个不同源文件中的两个init按其所在源文件名称的英文字母顺序执行
最好不要让声明在同一个包中的两个不同源文件中的两个init函数存在依赖关系
加载代码包时，所有声明的包级变量都在此包中的任何一个init函数执行之前初始化完毕
*/

//package main
//
//import "fmt"
//
//func init() {
//	fmt.Println("hi", bob)
//}
//
//func main() {
//	fmt.Println("bye")
//}
//
//func init() {
//	fmt.Println("hello", smith)
//}
//
//func titleName(who string) string {
//	return "Mr ." + who
//}
//
//var bob, smith = titleName("Bob"), titleName("Smith")

// 使用完整引入声明的代码
// 必须使用format和random作为限定标识符的前缀
//package main
//
//import (
//	format "fmt"
//	random "math/rand"
//	"time"
//)
//
//func main() {
//	random.Seed(time.Now().UnixNano())
//	format.Print("一个随机数：", random.Uint32(), "\n")
//}

// 完整引入声明语句形式引入名importname可以是一个句点（.）这样的引用成为句点引入
package main

import (
	. "fmt"
	. "time"
)

func main() {
	Println("Current time", Now())
}
