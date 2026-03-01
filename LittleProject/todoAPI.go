// todo_api.go
package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/dgrijalva/jwt-go"
)

type Task struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Status      string    `json:"status"`   // pending, in-progress, completed
	Priority    int       `json:"priority"` // 1-5
	Tags        []string  `json:"tags"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	DueDate     time.Time `json:"due_date"`
	UserID      string    `json:"user_id"`
}

type User struct {
	ID       string `json:"id"`
	Username string `json:"username"`
	Password string `json:"password"` // 实际应用中应该存储hash
	Email    string `json:"email"`
}

type TodoApp struct {
	mu        sync.RWMutex
	tasks     map[string]Task
	users     map[string]User
	jwtSecret []byte
}

func NewTodoApp() *TodoApp {
	return &TodoApp{
		tasks:     make(map[string]Task),
		users:     make(map[string]User),
		jwtSecret: []byte("your-secret-key"),
	}
}

// 中间件：认证
func (app *TodoApp) authMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		token := r.Header.Get("Authorization")
		if token == "" {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		// 移除 "Bearer " 前缀
		token = strings.TrimPrefix(token, "Bearer ")

		claims := jwt.MapClaims{}
		_, err := jwt.ParseWithClaims(token, claims, func(token *jwt.Token) (interface{}, error) {
			return app.jwtSecret, nil
		})

		if err != nil {
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}

		// 将用户ID添加到请求上下文
		r.Header.Set("X-User-ID", claims["user_id"].(string))
		next(w, r)
	}
}

// 创建任务
func (app *TodoApp) createTask(w http.ResponseWriter, r *http.Request) {
	var task Task
	if err := json.NewDecoder(r.Body).Decode(&task); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	userID := r.Header.Get("X-User-ID")
	task.ID = fmt.Sprintf("%d", time.Now().UnixNano())
	task.CreatedAt = time.Now()
	task.UpdatedAt = time.Now()
	task.UserID = userID

	if task.Status == "" {
		task.Status = "pending"
	}

	app.mu.Lock()
	app.tasks[task.ID] = task
	app.mu.Unlock()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(task)
}

// 获取任务列表
func (app *TodoApp) listTasks(w http.ResponseWriter, r *http.Request) {
	userID := r.Header.Get("X-User-ID")
	status := r.URL.Query().Get("status")
	tag := r.URL.Query().Get("tag")

	app.mu.RLock()
	defer app.mu.RUnlock()

	var tasks []Task
	for _, task := range app.tasks {
		if task.UserID != userID {
			continue
		}

		if status != "" && task.Status != status {
			continue
		}

		if tag != "" {
			found := false
			for _, t := range task.Tags {
				if t == tag {
					found = true
					break
				}
			}
			if !found {
				continue
			}
		}

		tasks = append(tasks, task)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tasks)
}

// 获取单个任务
func (app *TodoApp) getTask(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]
	userID := r.Header.Get("X-User-ID")

	app.mu.RLock()
	task, exists := app.tasks[id]
	app.mu.RUnlock()

	if !exists || task.UserID != userID {
		http.Error(w, "Task not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(task)
}

// 更新任务
func (app *TodoApp) updateTask(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]
	userID := r.Header.Get("X-User-ID")

	app.mu.Lock()
	task, exists := app.tasks[id]
	if !exists || task.UserID != userID {
		app.mu.Unlock()
		http.Error(w, "Task not found", http.StatusNotFound)
		return
	}

	var updates Task
	if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
		app.mu.Unlock()
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if updates.Title != "" {
		task.Title = updates.Title
	}
	if updates.Description != "" {
		task.Description = updates.Description
	}
	if updates.Status != "" {
		task.Status = updates.Status
	}
	if updates.Priority > 0 {
		task.Priority = updates.Priority
	}
	if updates.Tags != nil {
		task.Tags = updates.Tags
	}
	if !updates.DueDate.IsZero() {
		task.DueDate = updates.DueDate
	}
	task.UpdatedAt = time.Now()

	app.tasks[id] = task
	app.mu.Unlock()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(task)
}

// 删除任务
func (app *TodoApp) deleteTask(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id := vars["id"]
	userID := r.Header.Get("X-User-ID")

	app.mu.Lock()
	task, exists := app.tasks[id]
	if exists && task.UserID == userID {
		delete(app.tasks, id)
	}
	app.mu.Unlock()

	if !exists || task.UserID != userID {
		http.Error(w, "Task not found", http.StatusNotFound)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// 批量操作
func (app *TodoApp) batchUpdate(w http.ResponseWriter, r *http.Request) {
	var operations []map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&operations); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	userID := r.Header.Get("X-User-ID")
	results := make([]map[string]interface{}, 0)

	app.mu.Lock()
	defer app.mu.Unlock()

	for _, op := range operations {
		opType := op["type"].(string)
		taskID := op["task_id"].(string)

		result := map[string]interface{}{
			"task_id": taskID,
			"success": false,
		}

		task, exists := app.tasks[taskID]
		if !exists || task.UserID != userID {
			result["error"] = "Task not found"
			results = append(results, result)
			continue
		}

		switch opType {
		case "complete":
			task.Status = "completed"
			task.UpdatedAt = time.Now()
			app.tasks[taskID] = task
			result["success"] = true

		case "delete":
			delete(app.tasks, taskID)
			result["success"] = true

		case "update_priority":
			if priority, ok := op["priority"].(float64); ok {
				task.Priority = int(priority)
				task.UpdatedAt = time.Now()
				app.tasks[taskID] = task
				result["success"] = true
			}
		}

		results = append(results, result)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(results)
}

// 统计信息
func (app *TodoApp) getStats(w http.ResponseWriter, r *http.Request) {
	userID := r.Header.Get("X-User-ID")

	app.mu.RLock()
	defer app.mu.RUnlock()

	stats := map[string]interface{}{
		"total":       0,
		"pending":     0,
		"in_progress": 0,
		"completed":   0,
		"overdue":     0,
		"by_priority": make(map[int]int),
	}

	now := time.Now()
	for _, task := range app.tasks {
		if task.UserID != userID {
			continue
		}

		stats["total"] = stats["total"].(int) + 1
		stats[task.Status] = stats[task.Status].(int) + 1
		stats["by_priority"].(map[int]int)[task.Priority]++

		if task.Status != "completed" && !task.DueDate.IsZero() && task.DueDate.Before(now) {
			stats["overdue"] = stats["overdue"].(int) + 1
		}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}

func main() {
	app := NewTodoApp()
	router := mux.NewRouter()

	// API路由
	api := router.PathPrefix("/api").Subrouter()

	// 任务相关路由
	api.HandleFunc("/tasks", app.authMiddleware(app.createTask)).Methods("POST")
	api.HandleFunc("/tasks", app.authMiddleware(app.listTasks)).Methods("GET")
	api.HandleFunc("/tasks/{id}", app.authMiddleware(app.getTask)).Methods("GET")
	api.HandleFunc("/tasks/{id}", app.authMiddleware(app.updateTask)).Methods("PUT")
	api.HandleFunc("/tasks/{id}", app.authMiddleware(app.deleteTask)).Methods("DELETE")
	api.HandleFunc("/tasks/batch", app.authMiddleware(app.batchUpdate)).Methods("POST")
	api.HandleFunc("/stats", app.authMiddleware(app.getStats)).Methods("GET")

	log.Println("Todo API server starting on :8080")
	log.Fatal(http.ListenAndServe(":8080", router))
}
