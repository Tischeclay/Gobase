//package main
//
//import "fmt"
//
//func main() {
//	pm := &map[string]int{"C": 1972, "Go": 2009}
//	ps := &[]string{"break", "continue"}
//	pa := &[...]bool{false, true, true, false}
//	fmt.Printf("%T\n", pm) // *map[string]int
//	fmt.Printf("%T\n", ps) // *[]string
//	fmt.Printf("%T\n", pa) // *[4]bool
//}

package main

import "fmt"

func main() {
	var a [16]byte
	var s []int
	var m map[string]int

	fmt.Println(a == a)   // true
	fmt.Println(m == nil) // true
	fmt.Println(s == nil) // true
	fmt.Println(nil == map[string]int{}) // false
	fmt.Println(nil == []int{})          // false
