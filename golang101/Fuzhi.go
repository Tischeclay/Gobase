package main

import "fmt"

func main() {
	f := func(n int) int {
		fmt.Printf("f(%v) is called.\n", n)
		return n
	}

	switch x := f(3); x + f(4) {
	default:
	case f(5):
	case f(6), f(7), f(8):
	case f(9), f(10):
	}
}
