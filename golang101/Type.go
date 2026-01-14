/*
组合类型：
指针类型：类C指针
结构体类型：类C结构体
函数类型：在Go中是一等公民
容器类型：数组类型（定长容器类型）、切片类型（动态长度和容量容器类型）、映射类型（字典类型，在标准编译器中是使用哈希表实现的）
通道类型：通道用来同步并发的协程
接口类型：用于实现反射和多态
每一个基本类型和组合类型都对应一个类型种类
定义新的类型：
type NewTypeName SourceType
type (
	NewTypeName1 SourceType1
	NewTypeName2 SourceType2
)
新的类型名必须为标识符，但是包级类型名称不能为Init
类型别名声明
type (
	Name = string
	Age = int
)
type table = map[string]int
type Table = map[Name]Age
类型别名也必须为标识符，类型别名也可被声明在函数体内
底层类型：
每个类型都有一个底层类型，一个内置类型的底层类型是它自己，一个无名类型（组合类型）的底层类型为它自己，在类型声明中新声明的类型和源类型共享底层类型
*/
// 这四个类型的底层类型均为内置类型int。

package main

import "fmt"

func main() {
	fmt.Println("type (\nMyInt int\nAge   MyInt\n)\n\n// 下面这三个新声明的类型的底层类型各不相同。\ntype (\nIntSlice   []int   // 底层类型为[]int\nMyIntSlice []MyInt // 底层类型为[]MyInt\nAgeSlice   []Age   // 底层类型为[]Age\n)\n\n// 类型[]Age、Ages和AgeSlice的底层类型均为[]Age。\ntype Ages AgeSlice")
}
