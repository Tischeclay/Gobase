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

// filterAvailable 返回所有未借出的图书。
func filterAvailable(books []Book) []Book {
	var available []Book

	for _, book := range books {
		if !book.Borrowed {
			available = append(available, book)
		}
	}

	return available
}

// printBooks 打印图书列表。
func printBooks(books []Book) {
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
	printBooks(books)

	// 筛选并打印可借图书。
	available := filterAvailable(books)
	fmt.Printf("\n可借图书数量：%d\n", len(available))
	printBooks(available)

	// 通过 Library 方法查找图书。
	fmt.Println("\n查找 B002：")
	book, found := library.Find("B002")
	if found {
		fmt.Printf("找到：%s | %s\n", book.ID, book.Title)
	} else {
		fmt.Println("没有找到这本书")
	}

	// 查找不存在的编号。
	fmt.Println("\n查找 B999：")
	if _, found := library.Find("B999"); !found {
		fmt.Println("没有找到这本书")
	}

	// 演示：Find 返回的是图书副本，改副本不会影响书库中的原图书。
	fmt.Println("\n修改查找结果的副本：")
	if book, found := library.Find("B002"); found {
		book.Title = "修改后的书名"
		fmt.Println("副本的书名：", book.Title)

		original, _ := library.Find("B002")
		fmt.Println("书库中的书名：", original.Title)
	}

	// 借出 B001。
	fmt.Println("\n第一次借出 B001：")
	if err := library.Borrow("B001"); err != nil {
		fmt.Println("借出失败：", err)
	} else {
		fmt.Println("借出成功")
	}
	printBooks(books)

	// 再次借出同一本书，应返回错误。
	fmt.Println("\n再次借出 B001：")
	if err := library.Borrow("B001"); err != nil {
		fmt.Println("借出失败：", err)
	}

	// 借出不存在的图书，应返回错误。
	fmt.Println("\n借出 B999：")
	if err := library.Borrow("B999"); err != nil {
		fmt.Println("借出失败：", err)
	}

	// 归还 B001。
	fmt.Println("\n归还 B001：")
	if err := library.Return("B001"); err != nil {
		fmt.Println("归还失败：", err)
	} else {
		fmt.Println("归还成功")
	}
	printBooks(books)

	// 再次归还同一本书，应返回错误。
	fmt.Println("\n再次归还 B001：")
	if err := library.Return("B001"); err != nil {
		fmt.Println("归还失败：", err)
	}
}
