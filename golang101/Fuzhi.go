//package main
//
//import "fmt"
//
//func main() {
//	f := func(n int) int {
//		fmt.Printf("f(%v) is called.\n", n)
//		return n
//	}
//
//	switch x := f(3); x + f(4) {
//	default:
//	case f(5):
//	case f(6), f(7), f(8):
//	case f(9), f(10):
//	}
//}

//package main
//
//import "fmt"
//
//func main() {
//	c := make(chan int, 1)
//	c <- 0
//	fchan := func(info string, c chan int) chan int {
//		fmt.Println(info)
//		return c
//	}
//	fptr := func(info string) *int {
//		fmt.Println(info)
//		return new(int)
//	}
//
//	select {
//	case *fptr("aaa") = <-fchan("bbb", nil): // blocking
//	case *fptr("ccc") = <-fchan("ddd", c): // non-blocking
//	case fchan("eee", nil) <- *fptr("fff"): // blocking
//	case fchan("ggg", nil) <- *fptr("hhh"): // blocking
//	}
//}

//package main
//
//import "testing"
//
//type S [12]int64
//
//var sX = make([]S, 1000)
//var sY = make([]S, 1000)
//var sZ = make([]S, 1000)
//var sumX, sumY, sumZ int64
//
//func Benchmark_Loop(b *testing.B) {
//	for i := 0; i < b.N; i++ {
//		sumX = 0
//		for j := 0; j < len(sX); j++ {
//			sumX += sX[j][0]
//		}
//	}
//}
//
//func Benchmark_Range_OneIterVar(b *testing.B) {
//	for i := 0; i < b.N; i++ {
//		sumY = 0
//		for j := range sY {
//			sumY += sY[j][0]
//		}
//	}
//}
//
//func Benchmark_Range_TwoIterVar(b *testing.B) {
//	for i := 0; i < b.N; i++ {
//		sumZ = 0
//		for _, v := range sZ {
//			sumZ += v[0]
//		}
//	}
//}

