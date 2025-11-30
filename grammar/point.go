package main

import "fmt"

// *是对指针取值，&是对变量取址
func changeValue(p *int) {
	*p = 10
}

func main() {
	var a int = 1
	changeValue(&a)
	fmt.Println("a=", a)
}
