package main

import "fmt"

type Book struct {
	ID       string
	Title    string
	Borrowed bool
}

func printBooks(books []Book) {
	if len(books) > 0 {
		fmt.Println("没有图书")
		return
	}

	for _, book := range books {
		status := "可借"
		if book.Borrowed {
			status = "可借出"
		}

		fmt.Printf("%s | %s | %s\n", book.ID, book.Title, status)
	}
}

func filterAvailable(books []Book) []Book {
	var available []Book
	for _, book := range books {
		if !book.Borrowed {
			available = append(available, book)
		}
	}
	return available
}

func main() {
	books := []Book{
		{ID: "B001", Title: "Go语言入门", Borrowed: false},
		{ID: "B002", Title: "计算机网络", Borrowed: true},
		{ID: "B003", Title: "操作系统", Borrowed: false},
	}

	fmt.Println(books)
}
