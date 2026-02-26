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

//package main
//
//import (
//	"fmt"
//	"math/rand"
//	"time"
//)
//
//func source(c chan<- int32) {
//	ra, rb := rand.Int31(), rand.Intn(3)+1
//	// 睡眠1秒/2秒/3秒
//	time.Sleep(time.Duration(rb) * time.Second)
//	c <- ra
//}

//func main() {
//	rand.Seed(time.Now().UnixNano()) // Go 1.20之前需要
//
//	startTime := time.Now()
//	c := make(chan int32, 5) // 必须用一个缓冲通道
//	for i := 0; i < cap(c); i++ {
//		go source(c)
//	}
//	rnd := <-c // 只有第一个回应被使用了
//	fmt.Println(time.Since(startTime))
//	fmt.Println(rnd)
//}

//package main
//
//import (
//	"crypto/rand"
//	"fmt"
//	"os"
//	"sort"
//)
//
//func main() {
//	values := make([]byte, 32*1024*1024)
//	if _, err := rand.Read(values); err != nil {
//		fmt.Println(err)
//		os.Exit(1)
//	}
//
//	done := make(chan struct{}) // 也可以是缓冲的
//
//	// 排序协程
//	go func() {
//		sort.Slice(values, func(i, j int) bool {
//			return values[i] < values[j]
//		})
//		done <- struct{}{} // 通知排序已完成
//	}()
//
//	// 并发地做一些其它事情...
//
//	<-done // 等待通知
//	fmt.Println(values[0], values[len(values)-1])
//}

//package main
//
//import "fmt"
//
//type T int
//
//func IsClosed(ch <-chan T) bool {
//	select {
//	case <-ch:
//		return true
//	default:
//	}
//	return false
//}
//
//func main() {
//	c := make(chan T)
//	fmt.Println(IsClosed(c)) // false
//	close(c)
//	fmt.Println(IsClosed(c)) //true
//}

package main

import (
	"log"
	"math/rand"
	"sync"
	"time"
)

func main() {
	rand.Seed(time.Now().UnixNano()) // Go 1.20之前需要
	log.SetFlags(0)

	// ...
	const Max = 100000
	const NumReceivers = 100

	wgReceivers := sync.WaitGroup{}
	wgReceivers.Add(NumReceivers)

	// ...
	dataCh := make(chan int)

	// 发送者
	go func() {
		for {
			if value := rand.Intn(Max); value == 0 {
				// 此唯一的发送者可以安全地关闭此数据通道。
				close(dataCh)
				return
			} else {
				dataCh <- value
			}
		}
	}()

	// 接收者
	for i := 0; i < NumReceivers; i++ {
		go func() {
			defer wgReceivers.Done()

			// 接收数据直到通道dataCh已关闭
			// 并且dataCh的缓冲队列已空。
			for value := range dataCh {
				log.Println(value)
			}
		}()
	}

	wgReceivers.Wait()
}
