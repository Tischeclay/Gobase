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

// 这里使用sync标准库包中的WaitGroup来同步上面这个程序中的主协程和两个新创建的协程，WaitGroup类型有三个方法，Add方法
// 来注册新的需要完成的任务数，Done方法用来通知某个任务已经完成，Wait方法调用将阻塞（等待）到所有任务都已经完成后才继续执行其后的语句
package main

import (
	"log"
	"math/rand"
	"sync"
	"time"
)

var wg sync.WaitGroup

func SayGreetings(greeting string, times int) {
	for i := 0; i < times; i++ {
		log.Println(greeting)
		d := time.Second * time.Duration(rand.Intn(5)) / 2
		time.Sleep(d)
	}
	// 通知当前任务已完成
	wg.Done()
}

func main() {
	// Go 1.20之前需要
	rand.Seed(time.Now().UnixNano())
	log.SetFlags(0)
	// 注册两个新任务
	wg.Add(2)
	go SayGreetings("Hi", 10)
	go SayGreetings("Hello", 10)
	// 阻塞直至所有任务都已完成
	wg.Wait()
}
