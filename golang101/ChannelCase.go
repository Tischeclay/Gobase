//package main
//
//import (
//	"fmt"
//	"math/rand"
//	"time"
//)
//
//func longTimeRequest() <-chan int32 {
//	r := make(chan int32)
//
//	go func() {
//		time.Sleep(time.Second * 3) // 模拟一个工作负载
//		r <- rand.Int31n(100)
//	}()
//
//	return r
//}
//
//func sumSquares(a, b int32) int32 {
//	return a*a + b*b
//}
//
//func main() {
//	rand.Seed(time.Now().UnixNano()) // Go 1.20之前需要
//
//	a, b := longTimeRequest(), longTimeRequest()
//	fmt.Println(sumSquares(<-a, <-b))
//}

//package main
//
//import (
//	"fmt"
//	"math/rand"
//	"time"
//)
//
//func longTimeRequest(r chan<- int32) {
//	time.Sleep(time.Second * 3) // 模拟一个工作负载
//	r <- rand.Int31n(100)
//}
//
//func sumSquares(a, b int32) int32 {
//	return a*a + b*b
//}
//
//func main() {
//	rand.Seed(time.Now().UnixNano()) // Go 1.20之前需要
//
//	ra, rb := make(chan int32), make(chan int32)
//	go longTimeRequest(ra)
//	go longTimeRequest(rb)
//
//	fmt.Println(sumSquares(<-ra, <-rb))
//}

package main

import (
	"fmt"
	"math/rand"
	"time"
)

func source(c chan<- int32) {
	ra, rb := rand.Int31(), rand.Intn(3)+1
	// 睡眠1秒/2秒/3秒
	time.Sleep(time.Duration(rb) * time.Second)
	c <- ra
}

func main() {
	rand.Seed(time.Now().UnixNano()) // Go 1.20之前需要

	startTime := time.Now()
	c := make(chan int32, 5) // 必须用一个缓冲通道
	for i := 0; i < cap(c); i++ {
		go source(c)
	}
	rnd := <-c // 只有第一个回应被使用了
	fmt.Println(time.Since(startTime))
	fmt.Println(rnd)
}
