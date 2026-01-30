//package main
//
//import (
//	"fmt"
//	"time"
//)
//
//func main() {
//	c := make(chan int) // 一个非缓冲通道
//	go func(ch chan<- int, x int) {
//		time.Sleep(time.Second)
//		// <-ch    // 此操作编译不通过
//		ch <- x * x // 阻塞在此，直到发送的值被接收
//	}(c, 3)
//	done := make(chan struct{})
//	go func(ch <-chan int) {
//		n := <-ch      // 阻塞在此，直到有值发送到c
//		fmt.Println(n) // 9
//		// ch <- 123   // 此操作编译不通过
//		time.Sleep(time.Second)
//		done <- struct{}{}
//	}(c)
//	<-done // 阻塞在此，直到有值发送到done
//	fmt.Println("bye")
//}

//package main
//
//import "fmt"
//
//func main() {
//	c := make(chan int, 2) // 一个容量为2的缓冲通道
//	c <- 3
//	c <- 5
//	close(c)
//	fmt.Println(len(c), cap(c)) // 2 2
//	x, ok := <-c
//	fmt.Println(x, ok)          // 3 true
//	fmt.Println(len(c), cap(c)) // 1 2
//	x, ok = <-c
//	fmt.Println(x, ok)          // 5 true
//	fmt.Println(len(c), cap(c)) // 0 2
//	x, ok = <-c
//	fmt.Println(x, ok) // 0 false
//	x, ok = <-c
//	fmt.Println(x, ok)          // 0 false
//	fmt.Println(len(c), cap(c)) // 0 2
//	close(c)                    // 此行将产生一个恐慌
//	c <- 7                      // 如果上一行不存在，此行也将产生一个恐慌。
//}

//package main
//
//import (
//	"fmt"
//	"time"
//)
//
//func main() {
//	var ball = make(chan string)
//	kickBall := func(playerName string) {
//		for {
//			fmt.Print(<-ball, "传球", "\n")
//			time.Sleep(time.Second)
//			ball <- playerName
//		}
//	}
//	go kickBall("张三")
//	go kickBall("李四")
//	go kickBall("王二麻子")
//	go kickBall("刘大")
//	ball <- "裁判"    // 开球
//	var c chan bool // 一个零值nil通道
//	<-c             // 永久阻塞在此
//}

//package main
//
//import (
//	"fmt"
//	"time"
//)
//
//func main() {
//	fibonacci := func() chan uint64 {
//		c := make(chan uint64)
//		go func() {
//			var x, y uint64 = 0, 1
//			for ; y < (1 << 63); c <- y {
//				x, y = y, x+y
//			}
//			close(c)
//		}()
//		return c
//	}
//	c := fibonacci()
//	for x, ok := <-c; ok; x, ok = <-c {
//		time.Sleep(time.Second)
//		fmt.Println(x)
//	}
//}

// 一个不含任何分支的select-case代码块select{}将使当前协程处于永久阻塞状态。
// 在下面这个例子中，default分支将铁定得到执行，因为两个case分支后的操作均为阻塞的。
//package main
//
//import "fmt"
//
//func main() {
//	var c chan struct{}
//	select {
//	case <-c:
//	case c <- struct{}{}:
//	default:
//		fmt.Println("Go here.")
//	}
//}

//package main
//
//import "fmt"
//
//func main() {
//	c := make(chan string, 2)
//	trySend := func(v string) {
//		select {
//		case c <- v:
//		default: // 如果c的缓冲已满，则执行默认分支。
//		}
//	}
//	tryReceive := func() string {
//		select {
//		case v := <-c:
//			return v
//		default:
//			return "-" // 如果c的缓冲为空，则执行默认分支。
//		}
//	}
//	trySend("Hello!") // 发送成功
//	trySend("Hi!")    // 发送成功
//	trySend("Bye!")   // 发送失败，但不会阻塞。
//	// 下面这两行将接收成功。
//	fmt.Println(tryReceive()) // Hello!
//	fmt.Println(tryReceive()) // Hi!
//	// 下面这行将接收失败。
//	fmt.Println(tryReceive()) // -
//}

package main

func main() {
	c := make(chan struct{})
	close(c)
	select {
	case c <- struct{}{}: // 若此分支被选中，则产生一个恐慌
	case <-c:
	}
}
