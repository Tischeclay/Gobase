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

		fmt.Printf("%s | %s | %s\n",
			book.ID, book.Title, status)
	}
}

func findBook(books []Book, id string) (Book, bool) {
	for _, book := range books {
		if book.ID == id {
			return book, true
		}
	}

	// 必须遍历完所有图书，才能确定没有找到。
	return Book{}, false
}

// borrowBook 按编号借出图书。
// 成功返回 nil；不存在或已经借出时返回错误。
func borrowBook(books []Book, id string) error {
	for i := range books {
		// 不是要找的图书，继续检查下一本。
		if books[i].ID != id {
			continue
		}

		// 找到了，但已经借出，不能重复借。
		if books[i].Borrowed {
			return fmt.Errorf("图书 %s 已经借出", id)
		}

		// 通过索引修改原切片中的图书。
		books[i].Borrowed = true
		return nil
	}

	// 遍历结束仍未找到。
	return fmt.Errorf("图书 %s 不存在", id)
}

func main() {
	books := []Book{
		{ID: "B001", Title: "Go语言入门", Borrowed: false},
		{ID: "B002", Title: "计算机网络", Borrowed: true},
		{ID: "B003", Title: "操作系统", Borrowed: false},
	}

	// 昨天的功能：展示和筛选。
	fmt.Println("全部图书：")
	printBooks(books)

	available := filterAvailable(books)
	fmt.Printf("\n可借图书数量：%d\n", len(available))
	printBooks(available)

	// 今天增加的功能：按编号查找。
	fmt.Println("\n查找 B002：")
	book, found := findBook(books, "B002")
	if found {
		fmt.Printf("找到：%s | %s\n", book.ID, book.Title)
	} else {
		fmt.Println("没有找到这本书")
	}

	// 验证查找不存在的编号。
	fmt.Println("\n查找 B999：")
	if _, found := findBook(books, "B999"); !found {
		fmt.Println("没有找到这本书")
	}

	// 验证从空切片查找。
	fmt.Println("\n从空列表中查找：")
	if _, found := findBook([]Book{}, "B001"); !found {
		fmt.Println("没有找到这本书")
	}

	// 值复制实验：修改查找结果，不影响原切片里的图书。
	fmt.Println("\n修改返回的图书副本：")
	if book, found := findBook(books, "B002"); found {
		book.Title = "修改后的书名"
		fmt.Println("副本的书名：", book.Title)

		original, _ := findBook(books, "B002")
		fmt.Println("原来的书名：", original.Title)
	}
}
