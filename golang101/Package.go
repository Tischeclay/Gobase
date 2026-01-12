/*
内置函数print和println提供了和fmt标准库包中的对应函数相似的功能。 内置函数可以不用引入任何代码包而直接使用。
注意：print和println这两个内置函数不推荐使用在生产环境，因为它们不保证一定会出现在以后的Go版本中。
*/
package main

import "fmt"

func main() {
	a, b := 123, "Go"
	fmt.Printf("a == %v == 0x%x, b = %s\n", a, a, b)
	fmt.Printf("type of a : %T, type of b : %T\n", a, b)
	fmt.Printf("1%% 50%% 99%%\n")
}
