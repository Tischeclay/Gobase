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
	"sync"
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

func example2BufferedChannel() {
	fmt.Println("\n=== 示例2: 带缓冲通道（异步通信）===")

	// 创建容量为3的缓冲通道
	ch := make(chan string, 3)

	// 发送数据（不会阻塞直到缓冲区满）
	ch <- "消息1"
	ch <- "消息2"
	ch <- "消息3"
	fmt.Println("已发送3条消息，缓冲区使用情况:")
	fmt.Printf("  长度: %d, 容量: %d\n", len(ch), cap(ch))

	// 尝试发送第4条消息（会阻塞）
	go func() {
		time.Sleep(1 * time.Second)
		fmt.Println("准备发送第4条消息...")
		ch <- "消息4"
		fmt.Println("第4条消息已发送")
	}()
	// 接收数据
	for i := 0; i < 4; i++ {
		msg := <-ch
		fmt.Printf("接收: %s (缓冲区剩余: %d)\n", msg, len(ch))
	}
}

func example3DirectionalChannel() {
	fmt.Println("\n===单向通道，限制操作方向")
	// 双向通道
	ch := make(chan int)

	// 只发送函数
	sendOnly := func(send chan<- int) {
		for i := 1; i < 5; i++ {
			send <- i
			fmt.Printf("发送: %d\n", i)
			time.Sleep(100 * time.Millisecond)
		}
		close(send)
	}

	// 只接收函数
	receiveOnly := func(recv <-chan int) {
		for value := range recv {
			fmt.Println("接收: %d\n", value)
		}
	}

	go sendOnly(ch)
	receiveOnly(ch)
}

func example4Select() {
	fmt.Println("\n=== 示例4: Select多路复用 ===")

	ch1 := make(chan string)
	ch2 := make(chan string)
	ch3 := make(chan string)

	// 模拟多个服务
	go func() {
		time.Sleep(1 * time.Second)
		ch1 <- "来自服务1的消息"
	}()

	go func() {
		time.Sleep(2 * time.Second)
		ch2 <- "来自服务2的消息"
	}()

	go func() {
		time.Sleep(500 * time.Millisecond)
		ch3 <- "来自服务3的消息"
	}()
	// 使用select同时等待多个通道
	for i := 0; i < 3; i++ {
		select {
		case msg1 := <-ch1:
			fmt.Printf("✅ 收到: %s\n", msg1)
		case msg2 := <-ch2:
			fmt.Printf("✅ 收到: %s\n", msg2)
		case msg3 := <-ch3:
			fmt.Printf("✅ 收到: %s\n", msg3)
		case <-time.After(3 * time.Second):
			fmt.Println("⏰ 超时")
		}
	}
}

func example5ProducerConsumer() {
	fmt.Println("\n=== 示例5: 生产者-消费者模式 ===")

	const (
		producerCount = 3
		consumerCount = 2
		bufferSize    = 5
	)

	// 任务通道
	tasks := make(chan int, bufferSize)
	results := make(chan int, bufferSize)
	done := make(chan bool)

	// 生产者
	var wg sync.WaitGroup
	for i := 0; i < producerCount; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for j := 1; j <= 5; j++ {
				task := id*100 + j
				tasks <- task
				fmt.Printf("生产者%d 生产: %d\n", id, task)
				time.Sleep(time.Duration(100) * time.Millisecond)
			}
		}(i)
	}

	// 消费者
	for i := 0; i < consumerCount; i++ {
		go func(id int) {
			for task := range tasks {
				// 模拟处理耗时
				time.Sleep(150 * time.Millisecond)
				result := task * 2
				results <- result
				fmt.Printf("  消费者%d 消费: %d -> %d\n", id, task, result)
			}
		}(i)
	}
	// 等待所有生产者完成
	go func() {
		wg.Wait()
		close(tasks) // 关闭任务通道
	}()

	// 收集结果
	go func() {
		for result := range results {
			fmt.Printf("结果收集: %d\n", result)
		}
		done <- true
	}()

	// 等待所有结果处理完成
	time.Sleep(5 * time.Second)
	close(results)
	<-done
}

func example6FanOutFanIn() {
	fmt.Println("\n=== 扇入扇出模式 ===")
	// 输入通道
	input := make(chan int)
	// 启动多个协程
	workers := 3
	outputs := make([]<-chan int, workers)
	for i := 0; i < workers; i++ {
		outputs[i] = worker(i, input)
	}
	// 合并结果（扇入）
	merged := merge(outputs...)

	// 发送任务
	go func() {
		for i := 1; i <= 10; i++ {
			input <- i
		}
		close(input)
	}()

	// 接收结果
	for result := range merged {
		fmt.Printf("最终结果: %d\n", result)
	}
}

// 工作协程
func worker(id int, input <-chan int) <-chan int {
	output := make(chan int)
	go func() {
		defer close(output)
		for value := range input {
			// 模拟处理
			time.Sleep(50 * time.Millisecond)
			result := value * (id + 1)
			fmt.Printf("Worker %d 处理 %d -> %d\n", id, value, result)
			output <- result
		}
	}()
	return output
}

// 合并多个通道
func merge(channels ...<-chan int) <-chan int {
	var wg sync.WaitGroup
	out := make(chan int)
	// 为每个通道启动一个协程
	for _, ch := range channels {
		wg.Add(1)
		go func(c <-chan int) {
			defer wg.Done()
			for value := range c {
				out <- value
			}
		}(ch)

	}
	// 等待所有通道关闭后关闭输出通道
	go func() {
		wg.Wait()
		close(out)
	}()
	return out
}

func example8CloseChannel() {
	fmt.Println("\n=== 示例8: 关闭通道和range遍历 ===")

	ch := make(chan int)

	// 发送者
	go func() {
		for i := 1; i <= 5; i++ {
			ch <- i
			fmt.Printf("发送: %d\n", i)
		}
		close(ch) // 关闭通道
		fmt.Println("通道已关闭")
	}()

	// 使用range遍历（自动在通道关闭时退出）
	fmt.Println("开始接收:")
	for value := range ch {
		fmt.Printf("接收: %d\n", value)
	}
	fmt.Println("通道已关闭，range循环结束")

	// 检查通道是否关闭
	ch2 := make(chan int, 2)
	ch2 <- 1
	ch2 <- 2
	close(ch2)

	value, ok := <-ch2
	fmt.Printf("读取: %d, 通道状态: %v\n", value, ok)

	value, ok = <-ch2
	fmt.Printf("读取: %d, 通道状态: %v\n", value, ok)

	value, ok = <-ch2
	fmt.Printf("读取: %d, 通道状态: %v\n", value, ok)
}

func main() {
	exampleBasicChannel()
	example2BufferedChannel()
	example3DirectionalChannel()
	example4Select()
	example5ProducerConsumer()
	example6FanOutFanIn()
}
