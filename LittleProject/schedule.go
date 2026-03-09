// scheduler.go
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"
)

type Task struct {
	Name     string        `json:"name"`
	Command  string        `json:"command"`
	Args     []string      `json:"args"`
	Schedule string        `json:"schedule"` // cron格式: "* * * * *"
	Timeout  time.Duration `json:"timeout"`
	Enabled  bool          `json:"enabled"`
}

type Scheduler struct {
	tasks   map[string]*Task
	timers  map[string]*time.Timer
	done    chan bool
	mu      sync.RWMutex
	logFile *os.File
}

func NewScheduler(logPath string) (*Scheduler, error) {
	logFile, err := os.OpenFile(logPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, err
	}

	return &Scheduler{
		tasks:   make(map[string]*Task),
		timers:  make(map[string]*time.Timer),
		done:    make(chan bool),
		logFile: logFile,
	}, nil
}

func (s *Scheduler) AddTask(task *Task) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.tasks[task.Name]; exists {
		return fmt.Errorf("任务 %s 已存在", task.Name)
	}

	s.tasks[task.Name] = task
	return nil
}

func (s *Scheduler) Start() {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for name, task := range s.tasks {
		if task.Enabled {
			go s.scheduleTask(name, task)
		}
	}

	<-s.done
}

func (s *Scheduler) scheduleTask(name string, task *Task) {
	// 解析cron表达式（简化版）
	nextRun := s.parseCron(task.Schedule)

	timer := time.NewTimer(time.Until(nextRun))
	s.mu.Lock()
	s.timers[name] = timer
	s.mu.Unlock()

	<-timer.C

	// 执行任务
	s.executeTask(task)

	// 重新调度
	s.scheduleTask(name, task)
}

func (s *Scheduler) executeTask(task *Task) {
	log.Printf("执行任务: %s", task.Name)

	// 创建命令
	cmd := exec.Command(task.Command, task.Args...)

	// 设置超时
	if task.Timeout > 0 {
		timer := time.AfterFunc(task.Timeout, func() {
			cmd.Process.Kill()
			s.log("任务 %s 超时", task.Name)
		})
		defer timer.Stop()
	}

	// 执行
	output, err := cmd.CombinedOutput()

	// 记录结果
	if err != nil {
		s.log("任务 %s 失败: %v\n输出: %s", task.Name, err, string(output))
	} else {
		s.log("任务 %s 成功\n输出: %s", task.Name, string(output))
	}
}

func (s *Scheduler) parseCron(cron string) time.Time {
	// 简化实现：每分钟执行一次
	parts := strings.Split(cron, " ")
	if len(parts) != 5 {
		return time.Now().Add(1 * time.Minute)
	}

	// 这里应该实现真正的cron解析
	// 简单起见，返回下一分钟
	return time.Now().Add(1 * time.Minute)
}

func (s *Scheduler) log(format string, args ...interface{}) {
	timestamp := time.Now().Format("2006-01-02 15:04:05")
	msg := fmt.Sprintf("[%s] "+format+"\n", append([]interface{}{timestamp}, args...)...)

	// 写入文件
	s.logFile.WriteString(msg)

	// 同时输出到控制台
	fmt.Print(msg)
}

func (s *Scheduler) Stop() {
	s.mu.Lock()
	defer s.mu.Unlock()

	for name, timer := range s.timers {
		timer.Stop()
		delete(s.timers, name)
	}

	s.logFile.Close()
	close(s.done)
}

func loadTasks(filename string) ([]Task, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	var tasks []Task
	err = json.Unmarshal(data, &tasks)
	return tasks, err
}

func main() {
	configFile := flag.String("c", "tasks.json", "配置文件")
	logFile := flag.String("log", "scheduler.log", "日志文件")
	flag.Parse()

	// 加载任务
	tasks, err := loadTasks(*configFile)
	if err != nil {
		log.Fatal("加载任务失败:", err)
	}

	// 创建调度器
	scheduler, err := NewScheduler(*logFile)
	if err != nil {
		log.Fatal("创建调度器失败:", err)
	}
	defer scheduler.Stop()

	// 添加任务
	for _, task := range tasks {
		if err := scheduler.AddTask(&task); err != nil {
			log.Printf("添加任务 %s 失败: %v", task.Name, err)
		} else {
			log.Printf("添加任务: %s", task.Name)
		}
	}

	// 捕获退出信号
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		<-sigChan
		log.Println("收到退出信号，正在停止...")
		scheduler.Stop()
	}()

	// 启动调度器
	log.Println("调度器启动")
	scheduler.Start()
}

// tasks.json 示例:
// [
//   {
//     "name": "备份数据库",
//     "command": "mysqldump",
//     "args": ["-u", "root", "mydb", ">", "backup.sql"],
//     "schedule": "0 2 * * *",
//     "timeout": 3600000000000,
//     "enabled": true
//   }
// ]
