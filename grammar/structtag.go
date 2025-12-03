//package main
//
//import (
//	"fmt"
//	"reflect"
//)
//
//type resume struct {
//	Name   string `info:"name" doc:"my name"`
//	Gender string `info:"gender"`
//}
//
//func findTag(str interface{}) {
//	t := reflect.TypeOf(str).Elem()
//	for i := 0; i < t.NumField(); i++ {
//		taginfo := t.Field(i).Tag.Get("info")
//		tagdoc := t.Field(i).Tag.Get("doc")
//		fmt.Println("info: ", taginfo, " doc: ", tagdoc)
//	}
//}
//
//func main() {
//	var re resume
//	findTag(&re)
//}

package main

import (
	"encoding/json"
	"fmt"
)

type Movie struct {
	Title  string   `json:"title"`
	Year   int      `json:"year"`
	Price  float32  `json:"price"`
	Actors []string `json:"actors"`
}

func main() {
	movie := Movie{"喜剧之王", 2000, 10, []string{"Zhou Xingchi", "Zhang Baizhi"}}

	// 编码过程 结构体 -> json
	jsonStr, err := json.Marshal(movie)
	if err != nil {
		fmt.Println("json marshal error", err)
		return
	}
	fmt.Printf("jsonStr = %s\n", jsonStr)

	// 解码过程 jsonStr -> 结构体
	// jsonStr = {"喜剧之王", 2000, 10, []string{"Zhou Xingchi", "Zhang Baizhi"}}
	myMovie := Movie{}
	err = json.Unmarshal(jsonStr, &myMovie)
	if err != nil {
		fmt.Println("json unmarshal error", err)
		return
	}
	fmt.Printf("%v\n", myMovie)
}

//关于上述代码的疑惑
//// 1. Println: 把它当做 "uint8 类型的切片" -> 输出 ASCII 码数字
//fmt.Println("Println 输出:", jsonBytes)
//
//// 2. Printf + %v: 也是默认格式 -> 输出 ASCII 码数字
//fmt.Printf("Printf %%v 输出: %v\n", jsonBytes)
//
//// 3. Printf + %s: 把它当做 "文本" -> 输出 JSON 字符串
//fmt.Printf("Printf %%s 输出: %s\n", jsonBytes)
//
//// 4. 强制转换后用 Println: 先转成 string 类型，再打印 -> 输出 JSON 字符串
//fmt.Println("转 string 后输出:", string(jsonBytes))
