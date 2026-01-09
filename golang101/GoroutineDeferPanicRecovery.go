// 主协程创建两个新协程，time.Duration是一个在time标准库包中定义的类型，此类型的底层类型为内置类型int64
//package main
//
//import (
//	"log"
//	"math/rand"
//	"time"
//)
//
//func SayGreetings(greeting string, times int) {
//	for i := 0; i < times; i++ {
//		log.Println(greeting)
//		// 睡眠随机0-2.5s
//		d := time.Second * time.Duration(rand.Intn(5)) / 2
//		time.Sleep(d)
//	}
//}
//
//func main() {
//	rand.Seed(time.Now().UnixNano())
//	log.SetFlags(0)
//	go SayGreetings("Hi", 10)
//	go SayGreetings("Hello", 10)
//	time.Sleep(2 * time.Second)
//}

/* 这里使用sync标准库包中的WaitGroup来同步上面这个程序中的主协程和两个新创建的协程，WaitGroup类型有三个方法，Add方法
来注册新的需要完成的任务数，Done方法用来通知某个任务已经完成，Wait方法调用将阻塞（等待）到所有任务都已经完成后才继续执行其后的语句 */
//package main
//
//import (
//	"log"
//	"math/rand"
//	"sync"
//	"time"
//)
//
//var wg sync.WaitGroup
//
//func SayGreetings(greeting string, times int) {
//	for i := 0; i < times; i++ {
//		log.Println(greeting)
//		d := time.Second * time.Duration(rand.Intn(5)) / 2
//		time.Sleep(d)
//	}
//	// 通知当前任务已完成
//	wg.Done()
//}
//
//func main() {
//	// Go 1.20之前需要
//	rand.Seed(time.Now().UnixNano())
//	log.SetFlags(0)
//	// 注册两个新任务
//	wg.Add(2)
//	go SayGreetings("Hi", 10)
//	go SayGreetings("Hello", 10)
//	// 阻塞直至所有任务都已完成
//	wg.Wait()
//}

// Go协程的阻塞原理：协程只能从运行态被退出，处于阻塞状态的协程不会自发结束阻塞状态，当运行程序所有的协程都被阻塞时，程序进入死锁，官方编译器处理是让程序崩溃
//package main
//
//import (
//	"sync"
//	"time"
//)
//
//var wg sync.WaitGroup
//
//func main() {
//	wg.Add(1)
//	go func() {
//		time.Sleep(time.Second * 2)
//		// 阻塞
//		wg.Wait()
//	}()
//	// 阻塞
//	wg.Wait()
//}

/*
	当延迟调用语句被执行时，延迟函数调用会被推入由调用此延迟语句的函数调用维护一个延迟调用栈。当一个函数调用返回并进入其退出阶段后，栈中延迟函数逆序出栈并被执行

所有延迟调用执行完毕后函数调用就完全退出
*/
//package main
//
//import "fmt"
//
//func main() {
//	defer fmt.Println("the third line")
//	defer fmt.Println("the second line")
//	fmt.Println("the first line")
//}

//package main
//
//import "fmt"
//
//func main() {
//	defer fmt.Println("9")
//	fmt.Println("0")
//	defer fmt.Println("8")
//	fmt.Println("1")
//	if false {
//		defer fmt.Println("not reachable")
//	}
//	defer func() {
//		defer fmt.Println("7")
//		fmt.Println("3")
//		defer func() {
//			fmt.Println("5")
//			fmt.Println("6")
//		}()
//		fmt.Println("4")
//	}()
//	fmt.Println("2")
//	return
//	defer fmt.Println("not reachable")
//}

// 延迟调用可以修改包含此延迟调用最内层函数的返回值
//package main
//
//import "fmt"
//
//func Triple(n int) (r int) {
//	// 延迟调用函数defer func可以修改包含其最内层函数Triple的返回值
//	defer func() {
//		r += n
//	}()
//	return n + n
//}
//
//func main() {
//	fmt.Println(Triple(5))
//}

// eval-moment.go
/* 其中对于第二个匿名函数而言，Go 1.21和Go 1.22输出不同，这是由于Go1.22对for循环流程控制做出的语义修改导致代码变化导致的*/
//package main
//
//import "fmt"
//
//func main() {
//	func() {
//		var x = 0
//		for i := 0; i < 3; i++ {
//			defer fmt.Println("a:", i+x)
//		}
//		x = 10
//	}()
//	fmt.Println()
//	func() {
//		var x = 0
//		for i := 0; i < 3; i++ {
//			defer func() {
//				fmt.Println("b:", i+x)
//			}()
//		}
//		x = 10
//	}()
//	/*
//		for i := 0; i < 3; i++ {
//			i := 1
//			defer func() {
//				fmt.Println("b:", i)
//			}()
//		}()
//	*/
//}

// 在recover函数的返回值就是对应panic函数调用所消费的参数
// package main
//
// import "fmt"
//
//	func main() {
//		defer func() {
//			fmt.Println("正常退出")
//		}()
//		fmt.Println("嗨！")
//		defer func() {
//			v := recover()
//			fmt.Println("恐慌被恢复了:", v)
//		}()
//		panic("拜拜！")
//		fmt.Println("执行不到这里")
//	}
package main

import (
	"fmt"
	"time"
)

func main() {
	fmt.Println("hi!")
	go func() {
		time.Sleep(time.Second)
		panic(123)
	}()
	for {
		time.Sleep(time.Second)
	}
}
