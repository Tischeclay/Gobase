// filewatcher.go
package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/fsnotify/fsnotify"
)

type FileWatcher struct {
	watcher    *fsnotify.Watcher
	events     chan string
	errors     chan error
	watched    map[string]bool
	ignoreDirs []string
}

func NewFileWatcher() (*FileWatcher, error) {
	w, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}

	return &FileWatcher{
		watcher:    w,
		events:     make(chan string, 100),
		errors:     make(chan error, 10),
		watched:    make(map[string]bool),
		ignoreDirs: []string{".git", "node_modules", "vendor", "tmp"},
	}, nil
}

func (fw *FileWatcher) Watch(path string) error {
	err := filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// 检查是否应该忽略
		for _, ignore := range fw.ignoreDirs {
			if info.IsDir() && info.Name() == ignore {
				return filepath.SkipDir
			}
		}

		if info.IsDir() {
			if !fw.watched[path] {
				err = fw.watcher.Add(path)
				if err != nil {
					return err
				}
				fw.watched[path] = true
				log.Printf("监控目录: %s", path)
			}
		}
		return nil
	})

	return err
}

func (fw *FileWatcher) Start() {
	go func() {
		for {
			select {
			case event := <-fw.watcher.Events:
				fw.handleEvent(event)
			case err := <-fw.watcher.Errors:
				fw.errors <- err
			}
		}
	}()
}

func (fw *FileWatcher) handleEvent(event fsnotify.Event) {
	op := ""
	switch {
	case event.Op&fsnotify.Create == fsnotify.Create:
		op = "创建"
		// 如果是新创建的目录，自动监控
		if info, err := os.Stat(event.Name); err == nil && info.IsDir() {
			fw.watcher.Add(event.Name)
			fw.watched[event.Name] = true
		}
	case event.Op&fsnotify.Write == fsnotify.Write:
		op = "修改"
	case event.Op&fsnotify.Remove == fsnotify.Remove:
		op = "删除"
		delete(fw.watched, event.Name)
	case event.Op&fsnotify.Rename == fsnotify.Rename:
		op = "重命名"
		delete(fw.watched, event.Name)
	case event.Op&fsnotify.Chmod == fsnotify.Chmod:
		op = "权限变更"
	}

	if op != "" {
		fw.events <- fmt.Sprintf("[%s] %s: %s", time.Now().Format("15:04:05"), op, event.Name)
	}
}

func (fw *FileWatcher) Events() <-chan string {
	return fw.events
}

func (fw *FileWatcher) Errors() <-chan error {
	return fw.errors
}

func (fw *FileWatcher) Close() {
	fw.watcher.Close()
	close(fw.events)
	close(fw.errors)
}

func main() {
	if len(os.Args) < 2 {
		fmt.Printf("用法: %s <监控目录>\n", os.Args[0])
		os.Exit(1)
	}

	watchPath := os.Args[1]

	watcher, err := NewFileWatcher()
	if err != nil {
		log.Fatal(err)
	}
	defer watcher.Close()

	err = watcher.Watch(watchPath)
	if err != nil {
		log.Fatal(err)
	}

	watcher.Start()

	fmt.Printf("开始监控目录: %s\n", watchPath)
	fmt.Println("按 Ctrl+C 退出")

	for {
		select {
		case event := <-watcher.Events():
			fmt.Println(event)
		case err := <-watcher.Errors():
			log.Printf("错误: %v", err)
		}
	}
}
