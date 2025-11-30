package main

import "fmt"
import "math"

type Point struct{ X, Y float64 }

// 这是给struct Point类型定义一个方法
func (p Point) Distance(q Point) float64 {
	return math.Hypot(q.X-p.X, q.Y-p.Y)
}

func main() {

	p := Point{1, 2}
	q := Point{4, 6}

	distance1 := Point.Distance //方法表达式, 是一个函数值(相当于C语言的函数指针)
	fmt.Println(distance1(p, q))
	fmt.Printf("%T\n", distance1) //%T表示打出数据类型 ,这个必须放在Printf使用

	distance2 := (*Point).Distance //方法表达式,必须传递指针类型
	distance2(&p, q)
	fmt.Printf("%T\n", distance2)

}
