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

//package main
//
//import "fmt"
//import "reflect"
//
//type T []interface{ m() }
//
//func (T) m() {}
//
//func main() {
//	tp := reflect.TypeOf(new(interface{}))
//	tt := reflect.TypeOf(T{})
//	fmt.Println(tp.Kind(), tt.Kind()) // ptr slice
//
//	// 使用间接的方法得到表示两个接口类型的reflect.Type值。
//	ti, tim := tp.Elem(), tt.Elem()
//	fmt.Println(ti.Kind(), tim.Kind()) // interface interface
//
//	fmt.Println(tt.Implements(tim))  // true
//	fmt.Println(tp.Implements(tim))  // false
//	fmt.Println(tim.Implements(tim)) // true
//
//	// 所有的类型都实现了任何空接口类型。
//	fmt.Println(tp.Implements(ti))  // true
//	fmt.Println(tt.Implements(ti))  // true
//	fmt.Println(tim.Implements(ti)) // true
//	fmt.Println(ti.Implements(ti))  // true
//}

package main

import "fmt"
import "reflect"

type F func(string, int) bool

func (f F) m(s string) bool {
	return f(s, 32)
}
func (f F) M() {}

type I interface {
	m(s string) bool
	M()
}

func main() {
	var x struct {
		F F
		i I
	}
	tx := reflect.TypeOf(x)
	fmt.Println(tx.Kind())        // struct
	fmt.Println(tx.NumField())    // 2
	fmt.Println(tx.Field(1).Name) // i
	// 包路径（PkgPath）是非导出字段（或者方法）的内在属性。
	fmt.Println(tx.Field(0).PkgPath) //
	fmt.Println(tx.Field(1).PkgPath) // main

	tf, ti := tx.Field(0).Type, tx.Field(1).Type
	fmt.Println(tf.Kind())               // func
	fmt.Println(tf.IsVariadic())         // false
	fmt.Println(tf.NumIn(), tf.NumOut()) // 2 1
	t0, t1, t2 := tf.In(0), tf.In(1), tf.Out(0)
	// 下一行打印出：string int bool
	fmt.Println(t0.Kind(), t1.Kind(), t2.Kind())

	fmt.Println(tf.NumMethod(), ti.NumMethod()) // 1 2
	fmt.Println(tf.Method(0).Name)              // M
	fmt.Println(ti.Method(1).Name)              // m
	_, ok1 := tf.MethodByName("m")
	_, ok2 := ti.MethodByName("m")
	fmt.Println(ok1, ok2) // false true
}
