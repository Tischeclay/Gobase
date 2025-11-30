package main

import "fmt"

// 切片的传递实质上是引用传递而不是值传递，不同元素长度的动态数组形参是一致的
// 对于静态数组而言，数组声明时数组长度是固定的，固定长度的数组在传参的时候是严格匹配数组类型的，是值传递
func printArray(myArray []int) {
	for _, value := range myArray {
		fmt.Println("value = ", value)
	}
}

func main() {
	myArray := []int{1, 2, 3, 4}
	fmt.Println(myArray)
	printArray(myArray)
}
