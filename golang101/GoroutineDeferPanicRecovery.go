// 主协程创建两个新协程，time.Duration是一个在time标准库包中定义的类型，此类型的底层类型为内置类型int64
package main

import (
	"log"
	"math/rand"
	"time"
)

func SayGreetings(greeting string, times int) {
	for i := 0; i < times; i++ {
		log.Println(greeting)
		// 睡眠随机0-2.5s
		d := time.Second * time.Duration(rand.Intn(5)) / 2
		time.Sleep(d)
	}
}

func main() {
	rand.Seed(time.Now().UnixNano())
	log.SetFlags(0)
	go SayGreetings("Hi", 10)
	go SayGreetings("Hello", 10)
	time.Sleep(2 * time.Second)
}
