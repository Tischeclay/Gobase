//package main
//
//import "net/http"
//
//func main() {
//	type P = *bool
//	type M = map[int]int
//	var x struct {
//		string
//		error
//		*int
//		P
//		M
//		http.Header
//	}
//	x.string = "Go"
//	x.error = nil
//	x.int = new(int)
//	x.P = new(bool)
//	x.M = make(M)
//	x.Header = http.Header{}
//}

// 通过类型内嵌来扩展类型功能的例子
package main

import "fmt"

type Person struct {
	Name string
	Age  int
}

func (p Person) PrintName() {
	fmt.Println("Name", p.Name)
}

func (p *Person) SetAge(age int) {
	p.Age = age
}

type Singer struct {
	Person
	works []string
}

func main() {
	var gaga = Singer{Person: Person{"Gaga", 30}}
	gaga.PrintName()
	gaga.Name = "Polina Gagarina"
	(&gaga).SetAge(31)
	(&gaga).PrintName()
	fmt.Println(gaga.Age)
}
