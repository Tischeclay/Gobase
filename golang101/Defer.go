//package main
//
//import "fmt"
//
//func main() {
//	s := []string{"a", "b", "c", "d"}
//	defer fmt.Println(s) // [a x y d]
//	// defer append(s[:1], "x", "y") // 编译错误
//	defer func() {
//		_ = append(s[:1], "x", "y")
//	}()
//}

//package main
//
//import "fmt"
//
//func main() {
//	defer fmt.Println("此行可以被执行到")
//	var f func() // f == nil
//	defer f()    // 将产生一个恐慌
//	fmt.Println("此行可以被执行到")
//	f = func() {} // 此行不会阻止恐慌产生
//}

//package main
//
//type T int
//
//func (t T) M(n int) T {
//	print(n)
//	return t
//}
//
//func main() {
//	var t T
//	// t.M(1)是方法调用M(2)的属主实参，因此它
//	// 将在M(2)调用被推入延迟调用栈时被估值。
//	defer t.M(1).M(2)
//	t.M(3).M(4)
//}

//package main
//
//import "os"
//
//func withoutDefers(filepath string, head, body []byte) error {
//	f, err := os.Open(filepath)
//	if err != nil {
//		return err
//	}
//
//	_, err = f.Seek(16, 0)
//	if err != nil {
//		f.Close()
//		return err
//	}
//
//	_, err = f.Write(head)
//	if err != nil {
//		f.Close()
//		return err
//	}
//
//	_, err = f.Write(body)
//	if err != nil {
//		f.Close()
//		return err
//	}
//
//	err = f.Sync()
//	f.Close()
//	return err
//}
//
//func withDefers(filepath string, head, body []byte) error {
//	f, err := os.Open(filepath)
//	if err != nil {
//		return err
//	}
//	defer f.Close()
//
//	_, err = f.Seek(16, 0)
//	if err != nil {
//		return err
//	}
//
//	_, err = f.Write(head)
//	if err != nil {
//		return err
//	}
//
//	_, err = f.Write(body)
//	if err != nil {
//		return err
//	}
//
//	return f.Sync()
//}

var m sync.Mutex

func f1() {
m.Lock()
defer m.Unlock()
doSomething()
}

func f2() {
m.Lock()
doSomething()
m.Unlock()
}
func writeManyFiles(files []File) error {
for _, file := range files {
if err := func() error {
f, err := os.Open(file.path)
if err != nil {
return err
}
defer f.Close() // 将在此循环步内执行

_, err = f.WriteString(file.content)
if err != nil {
return err
}

return f.Sync()
}(); err != nil {
return err
}
}

return nil
}