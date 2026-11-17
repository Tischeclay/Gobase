package main

import "fmt"

// Book 表示一本图书。
type Book struct {
	ID       string
	Title    string
	Borrowed bool
}

// Library 保存图书列表。
type Library struct {
	books []Book
}

// NewLibrary 创建一个书库。
func NewLibrary(books []Book) *Library {
	return &Library{books: books}
}

// Find 按编号查找图书。
// 找到时返回图书副本和 true；找不到时返回零值图书和 false。
func (l *Library) Find(id string) (Book, bool) {
	for _, book := range l.books {
		if book.ID == id {
			return book, true
		}
	}
	return Book{}, false
}

// Available 返回所有未借出的图书。
func (l *Library) Available() []Book {
	var available []Book

	for _, book := range l.books {
		if !book.Borrowed {
			available = append(available, book)
		}
	}

	return available
}

// Borrow 按编号借出图书。
func (l *Library) Borrow(id string) error {
	for i := range l.books {
		if l.books[i].ID != id {
			continue
		}

		if l.books[i].Borrowed {
			return fmt.Errorf("图书 %s 已经借出", id)
		}

		l.books[i].Borrowed = true
		return nil
	}

	return fmt.Errorf("图书 %s 不存在", id)
}

// Return 按编号归还图书。
func (l *Library) Return(id string) error {
	for i := range l.books {
		if l.books[i].ID != id {
			continue
		}

		if !l.books[i].Borrowed {
			return fmt.Errorf("图书 %s 当前未借出", id)
		}

		l.books[i].Borrowed = false
		return nil
	}

	return fmt.Errorf("图书 %s 不存在", id)
}

// PrintBooks 打印传入的图书列表。
func (l *Library) PrintBooks(books []Book) {
	if len(books) == 0 {
		fmt.Println("没有图书")
		return
	}

	for _, book := range books {
		status := "可借"
		if book.Borrowed {
			status = "已借出"
		}

		fmt.Printf("%s | %s | %s\n", book.ID, book.Title, status)
	}
}

func main() {
	// 创建初始图书数据。
	books := []Book{
		{ID: "B001", Title: "Go语言入门"},
		{ID: "B002", Title: "计算机网络", Borrowed: true},
		{ID: "B003", Title: "操作系统"},
	}

	library := NewLibrary(books)

	fmt.Println("全部图书：")
	library.PrintBooks(books)

	// 筛选并打印可借图书。
	available := library.Available()
	fmt.Printf("\n可借图书数量：%d\n", len(available))
	library.PrintBooks(available)

	// 查找图书。
	fmt.Println("\n查找 B002：")
	book, found := library.Find("B002")
	if found {
		fmt.Printf("找到：%s | %s\n", book.ID, book.Title)
	} else {
		fmt.Println("没有找到这本书")
	}

	// 借出 B001。
	fmt.Println("\n借出 B001：")
	if err := library.Borrow("B001"); err != nil {
		fmt.Println("借出失败：", err)
	} else {
		fmt.Println("借出成功")
	}
	library.PrintBooks(books)

	// 再次借出同一本书，应失败。
	fmt.Println("\n再次借出 B001：")
	if err := library.Borrow("B001"); err != nil {
		fmt.Println("借出失败：", err)
	}

	// 归还 B001。
	fmt.Println("\n归还 B001：")
	if err := library.Return("B001"); err != nil {
		fmt.Println("归还失败：", err)
	} else {
		fmt.Println("归还成功")
	}
	library.PrintBooks(books)
}
