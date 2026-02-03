//package main
//
//import "fmt"
//import "reflect"
//
//func main() {
//	type A = [16]int16
//	var c <-chan map[A][]byte
//	tc := reflect.TypeOf(c)
//	fmt.Println(tc.Kind())    // chan
//	fmt.Println(tc.ChanDir()) // <-chan
//	tm := tc.Elem()
//	ta, tb := tm.Key(), tm.Elem()
//	fmt.Println(tm.Kind(), ta.Kind(), tb.Kind()) // map array slice
//	tx, ty := ta.Elem(), tb.Elem()
//
//	// byte是uint8类型的别名。
//	fmt.Println(tx.Kind(), ty.Kind()) // int16 uint8
//	fmt.Println(tx.Bits(), ty.Bits()) // 16 8
//	fmt.Println(tx.ConvertibleTo(ty)) // true
//	fmt.Println(tb.ConvertibleTo(ta)) // false
//
//	// 切片类型和映射类型都是不可比较类型。
//	fmt.Println(tb.Comparable()) // false
//	fmt.Println(tm.Comparable()) // false
//	fmt.Println(ta.Comparable()) // true
//	fmt.Println(tc.Comparable()) // true
//}

package main

import "fmt"
import "reflect"

type T []interface{ m() }

func (T) m() {}

func main() {
	tp := reflect.TypeOf(new(interface{}))
	tt := reflect.TypeOf(T{})
	fmt.Println(tp.Kind(), tt.Kind()) // ptr slice

	// 使用间接的方法得到表示两个接口类型的reflect.Type值。
	ti, tim := tp.Elem(), tt.Elem()
	fmt.Println(ti.Kind(), tim.Kind()) // interface interface

	fmt.Println(tt.Implements(tim))  // true
	fmt.Println(tp.Implements(tim))  // false
	fmt.Println(tim.Implements(tim)) // true

	// 所有的类型都实现了任何空接口类型。
	fmt.Println(tp.Implements(ti))  // true
	fmt.Println(tt.Implements(ti))  // true
	fmt.Println(tim.Implements(ti)) // true
	fmt.Println(ti.Implements(ti))  // true
}
