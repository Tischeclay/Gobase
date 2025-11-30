package main

import "fmt"

func main() {
	var a int
	fmt.Println("a=", a)
	fmt.Printf("type of b = %T\n", a)

	var bb string = "hello"
	fmt.Println("bb= %s, type of bb = %T\n", bb, bb)
}
