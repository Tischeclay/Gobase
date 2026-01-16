// 加入分号后被ide自动忽略并清除了
package main

import "fmt"

func main() {
	var (
		i   int
		sum int
	)
	for i < 6 {
		sum += i
		i++
	}
	fmt.Println(sum)
}
