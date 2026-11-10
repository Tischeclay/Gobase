package main

import "fmt"

type Book struct {
	ID       string
	Title    string
	Borrowed bool
}

func main() {
	books := []Book{
		{ID: "B001", Title: "Go语言入门", Borrowed: false},
		{ID: "B002", Title: "计算机网络", Borrowed: true},
		{ID: "B003", Title: "操作系统", Borrowed: false},
	}

	fmt.Println(books)
}
