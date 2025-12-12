package main

import "fmt"

func main() {
	c := make(chan int)
	go func() {
		defer fmt.Println("goroutine exit")
		fmt.Println("goroutine running")
		c <- 666
	}()

	num := <-c
	fmt.Println("num = ", num)
	fmt.Println("main goroutine exit")
}
