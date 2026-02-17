//package main
//
//import "fmt"
//
//func main() {
//	f := func(n int) int {
//		fmt.Printf("f(%v) is called.\n", n)
//		return n
//	}
//
//	switch x := f(3); x + f(4) {
//	default:
//	case f(5):
//	case f(6), f(7), f(8):
//	case f(9), f(10):
//	}
//}

package main

import "fmt"

func main() {
	c := make(chan int, 1)
	c <- 0
	fchan := func(info string, c chan int) chan int {
		fmt.Println(info)
		return c
	}
	fptr := func(info string) *int {
		fmt.Println(info)
		return new(int)
	}

	select {
	case *fptr("aaa") = <-fchan("bbb", nil): // blocking
	case *fptr("ccc") = <-fchan("ddd", c): // non-blocking
	case fchan("eee", nil) <- *fptr("fff"): // blocking
	case fchan("ggg", nil) <- *fptr("hhh"): // blocking
	}
}
