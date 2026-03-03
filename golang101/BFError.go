package main

import (
	"runtime"
	"time"
)

func main() {
	var a []int // nil
	var b bool  // false

	// 一个匿名协程。
	go func() {
		a = make([]int, 3)
		b = true // 写入b
	}()

	for !b { // 读取b
		time.Sleep(time.Second)
		runtime.Gosched()
	}
	a[0], a[1], a[2] = 0, 1, 2 // 可能会发生恐慌
}
