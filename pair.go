package main

import "fmt"

//func main() {
//	var a string
//	// pair<statictype:string, value:"a string">
//	a = "a string"
//
//	// pair<type:string, value:"a string">
//	var allType interface{}
//	allType = a
//
//	str, _ := allType.(string)
//	fmt.Println(str)
//}

//func main() {
//	// tty: pair<type:*os.File, value:"/dev/tty"文件描述符>
//	tty, err := os.OpenFile("/dev/tty", os.O_RDWR, 0)
//	if err != nil {
//		fmt.Println(err)
//		return
//	}
//
//	// r : pair<type: , value:>
//	var r io.Reader
//	// r: pair<type:*os.File, value:"/dev/tty"文件描述符>
//	r = tty
//
//	// w: pair<type: , value: >
//	var w io.Writer
//	// w: pair<type:*os.File, value:"/dev/tty"文件描述符>
//	w = r.(io.Writer)
//
//	w.Write([]byte("HELLO THIS is A TEST!!!\n"))
//}

type Reader interface {
	ReadBook()
}

type Writer interface {
	WriteBook()
}

type ReadWriter struct {
}

func (this *ReadWriter) ReadBook() {
	fmt.Println("ReadBook")
}

func (this *ReadWriter) WriteBook() {
	fmt.Println("WriteBook")
}

func main() {
	// raw: pair<type:ReadWriter, value:ReadWriter{}地址>
	raw := &ReadWriter{}
	// r: pair<type: , value: >
	var r Reader
	// r: pair<type:ReadWriter, value:ReadWriter{}地址>
	r = raw
	r.ReadBook()
	var w Writer
	w = r.(Writer)
	w.WriteBook()
}
