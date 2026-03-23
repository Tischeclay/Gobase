//package main
//
//import "fmt"
//
//func main() {
//	c := make(chan int)
//	go func() {
//		defer fmt.Println("goroutine exit")
//		fmt.Println("goroutine running")
//		c <- 666
//	}()
//
//	num := <-c
//	fmt.Println("num = ", num)
//	fmt.Println("main goroutine exit")
//}

// 无缓冲的channel
//package main
//
//import (
//	"fmt"
//	"time"
//)
//
//func main() {
//	c := make(chan int, 0) //创建无缓冲的通道 c
//
//	//内置函数 len 返回未被读取的缓冲元素数量，cap 返回缓冲区大小
//	fmt.Printf("len(c)=%d, cap(c)=%d\n", len(c), cap(c))
//
//	go func() {
//		defer fmt.Println("子go程结束")
//
//		for i := 0; i < 3; i++ {
//			c <- i
//			fmt.Printf("子go程正在运行[%d]: len(c)=%d, cap(c)=%d\n", i, len(c), cap(c))
//		}
//	}()
//
//	time.Sleep(2 * time.Second) //延时2s
//
//	for i := 0; i < 3; i++ {
//		num := <-c //从c中接收数据，并赋值给num
//		fmt.Println("num = ", num)
//	}
//
//	fmt.Println("main进程结束")
//}

package main

import (
	"fmt"
	"time"
)

func exampleBasicChannel() {
	fmt.Println("无缓冲通道，同步通信")
	// 创建一个无缓冲的int类型的通道
	ch := make(chan int)

	// 发送goroutine
	go func() {
		fmt.Println("发送goroutine：准备发送数据")
		time.Sleep(2 * time.Second)
		ch <- 42
		fmt.Println("发送goroutine：数据已发送")
	}()

	// 接收goroutine
	fmt.Println("主goroutine:等待接收数据")
	value := <-ch
	fmt.Printf("主goroutine:接收到数据%d\n", value)
}

func main() {
	exampleBasicChannel()
}
