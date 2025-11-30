package main

import "fmt"

func main() {
	// 声明方法一，使用前声明，make赋值
	var test1 map[string]string
	test1 = make(map[string]string, 10)
	test1["one"] = "php"
	test1["two"] = "java"
	test1["three"] = "python"
	fmt.Println(test1)

	// 声明方法二
	test2 := map[string]string{}
	test2["one"] = "php"
	test2["two"] = "java"
	fmt.Println(test2)

	// 声明方法三
	test3 := map[string]string{
		"one":   "php",
		"two":   "java",
		"three": "python",
	}
	fmt.Println(test3)

	language := make(map[string]map[string]string)
	language["php"] = make(map[string]string, 2)
	language["php"]["id"] = "1"
	language["php"]["desc"] = "php是世界上最美的语言"
	language["golang"] = make(map[string]string, 2)
	language["golang"]["id"] = "2"
	language["golang"]["desc"] = "golang抗并发非常good"
	fmt.Println(language) //map[php:map[id:1 desc:php是世界上最美的语言] golang:map[id:2 desc:golang抗并发非常good]]

	//增删改查
	val, key := language["php"] //查找是否有php这个子元素
	if key {
		fmt.Printf("%v", val)
	} else {
		fmt.Printf("no")
	}

	language["php"]["id"] = "3"         //修改了php子元素的id值
	language["php"]["nickname"] = "啪啪啪" //增加php元素里的nickname值
	delete(language, "php")             //删除了php子元素
	fmt.Println(language)
}
