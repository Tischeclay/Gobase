package main

import "fmt"

func main() {
	i := 0

Next: // 跳转标签声明
	fmt.Println(i)
	i++
	if i < 5 {
		goto Next // 跳转
	}
}
