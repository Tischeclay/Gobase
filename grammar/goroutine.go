package main

import (
	"fmt"
	"time"
)

//func newTask() {
//	i := 0
//	for {
//		i++
//		fmt.Printf("new Goroutine : i = %d\n", i)
//		time.Sleep(1 * time.Second)
//	}
//}
//
//func main() {
//	go newTask()
//	fmt.Println("main goroutine exit")
//}

func main() {
	go func() {
		defer fmt.Println("A.defer")

		func() {
			defer fmt.Println("B.defer")
			// 退出当前goroutine
			// runtime.Goexit()
			fmt.Println("B")
		}()
		fmt.Println("A")
	}()

	for {
		time.Sleep(1 * time.Second)
	}
}
