package main

import (
	"bufio"
	"fmt"
	"os"
	"sort"
	"strings"
)

// Book 表示一本图书。
type Book struct {
	ID       string
	Title    string
	Borrowed bool
}

// Library 保存图书。
type Library struct {
	books []Book
}

// NewLibrary 创建书库，并复制传入的图书切片。
func NewLibrary(books []Book) *Library {
	bookCopy := make([]Book, len(books))
	copy(bookCopy, books)

	return &Library{books: bookCopy}
}

// Add 添加图书。编号和书名不能为空，编号不能重复。
// 新书默认是未借出状态。
func (l *Library) Add(book Book) error {
	book.ID = strings.TrimSpace(book.ID)
	book.Title = strings.TrimSpace(book.Title)

	if book.ID == "" {
		return fmt.Errorf("图书编号不能为空")
	}
	if book.Title == "" {
		return fmt.Errorf("书名不能为空")
	}
	if _, found := l.Find(book.ID); found {
		return fmt.Errorf("图书编号 %s 已存在", book.ID)
	}

	book.Borrowed = false
	l.books = append(l.books, book)
	return nil
}

// List 返回按编号排序的图书副本。
func (l *Library) List() []Book {
	books := make([]Book, len(l.books))
	copy(books, l.books)

	sort.Slice(books, func(i, j int) bool {
		return books[i].ID < books[j].ID
	})

	return books
}

// Find 按编号查找图书。
func (l *Library) Find(id string) (Book, bool) {
	id = strings.TrimSpace(id)

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

	for _, book := range l.List() {
		if !book.Borrowed {
			available = append(available, book)
		}
	}

	return available
}

// Borrow 借出指定图书。
func (l *Library) Borrow(id string) error {
	index := l.findIndex(id)
	if index == -1 {
		return fmt.Errorf("图书 %s 不存在", id)
	}
	if l.books[index].Borrowed {
		return fmt.Errorf("图书 %s 已经借出", id)
	}

	l.books[index].Borrowed = true
	return nil
}

// Return 归还指定图书。
func (l *Library) Return(id string) error {
	index := l.findIndex(id)
	if index == -1 {
		return fmt.Errorf("图书 %s 不存在", id)
	}
	if !l.books[index].Borrowed {
		return fmt.Errorf("图书 %s 当前未借出", id)
	}

	l.books[index].Borrowed = false
	return nil
}

// Delete 删除未借出的图书。
func (l *Library) Delete(id string) error {
	index := l.findIndex(id)
	if index == -1 {
		return fmt.Errorf("图书 %s 不存在", id)
	}
	if l.books[index].Borrowed {
		return fmt.Errorf("图书 %s 已借出，不能删除", id)
	}

	l.books = append(l.books[:index], l.books[index+1:]...)
	return nil
}

// findIndex 查找图书在内部切片中的位置，找不到时返回 -1。
func (l *Library) findIndex(id string) int {
	id = strings.TrimSpace(id)

	for i := range l.books {
		if l.books[i].ID == id {
			return i
		}
	}

	return -1
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

// readInput 显示提示并读取一行输入。
func readInput(scanner *bufio.Scanner, prompt string) (string, bool) {
	fmt.Print(prompt)

	if !scanner.Scan() {
		return "", false
	}

	return strings.TrimSpace(scanner.Text()), true
}

func main() {
	library := NewLibrary([]Book{
		{ID: "B001", Title: "Go语言入门"},
		{ID: "B002", Title: "计算机网络", Borrowed: true},
		{ID: "B003", Title: "操作系统"},
	})

	scanner := bufio.NewScanner(os.Stdin)

	for {
		fmt.Println("\n===== 图书借阅系统 =====")
		fmt.Println("1. 查看全部图书")
		fmt.Println("2. 查看可借图书")
		fmt.Println("3. 查找图书")
		fmt.Println("4. 添加图书")
		fmt.Println("5. 借出图书")
		fmt.Println("6. 归还图书")
		fmt.Println("7. 删除图书")
		fmt.Println("0. 退出")

		choice, ok := readInput(scanner, "请选择：")
		if !ok {
			fmt.Println("\n输入结束，程序退出。")
			return
		}

		switch choice {
		case "1":
			printBooks(library.List())

		case "2":
			printBooks(library.Available())

		case "3":
			id, ok := readInput(scanner, "请输入图书编号：")
			if !ok {
				fmt.Println("\n输入结束，程序退出。")
				return
			}

			book, found := library.Find(id)
			if !found {
				fmt.Println("没有找到这本书")
				continue
			}

			printBooks([]Book{book})

		case "4":
			id, ok := readInput(scanner, "请输入图书编号：")
			if !ok {
				fmt.Println("\n输入结束，程序退出。")
				return
			}

			title, ok := readInput(scanner, "请输入书名：")
			if !ok {
				fmt.Println("\n输入结束，程序退出。")
				return
			}

			err := library.Add(Book{
				ID:    id,
				Title: title,
			})
			if err != nil {
				fmt.Println("添加失败：", err)
			} else {
				fmt.Println("添加成功")
			}

		case "5":
			id, ok := readInput(scanner, "请输入要借出的图书编号：")
			if !ok {
				fmt.Println("\n输入结束，程序退出。")
				return
			}

			if err := library.Borrow(id); err != nil {
				fmt.Println("借出失败：", err)
			} else {
				fmt.Println("借出成功")
			}

		case "6":
			id, ok := readInput(scanner, "请输入要归还的图书编号：")
			if !ok {
				fmt.Println("\n输入结束，程序退出。")
				return
			}

			if err := library.Return(id); err != nil {
				fmt.Println("归还失败：", err)
			} else {
				fmt.Println("归还成功")
			}

		case "7":
			id, ok := readInput(scanner, "请输入要删除的图书编号：")
			if !ok {
				fmt.Println("\n输入结束，程序退出。")
				return
			}

			if err := library.Delete(id); err != nil {
				fmt.Println("删除失败：", err)
			} else {
				fmt.Println("删除成功")
			}

		case "0":
			fmt.Println("再见！")
			return

		default:
			fmt.Println("无效选项，请输入 0～7。")
		}
	}
}
