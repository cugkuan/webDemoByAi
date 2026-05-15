package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
)

// Response 标准响应格式
type Response struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Data    interface{} `json:"data,omitempty"`
}

// Task 任务结构
type Task struct {
	ID    int    `json:"id"`
	Title string `json:"title"`
	Done  bool   `json:"done"`
}

var tasks = []Task{
	{ID: 1, Title: "学习 Go", Done: false},
	{ID: 2, Title: "构建 REST API", Done: true},
}

// getTasks 获取所有任务
func getTasks(w http.ResponseWriter, r *http.Request) {
	resp := Response{
		Code:    200,
		Message: "成功",
		Data:    tasks,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

// getTaskByID 获取单个任务
func getTaskByID(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(r.URL.Path, "/")
	id, err := strconv.Atoi(parts[len(parts)-1])
	if err != nil {
		resp := Response{Code: 400, Message: "无效的任务 ID"}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(resp)
		return
	}

	for _, task := range tasks {
		if task.ID == id {
			resp := Response{
				Code:    200,
				Message: "成功",
				Data:    task,
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(resp)
			return
		}
	}

	resp := Response{Code: 404, Message: "任务不存在"}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusNotFound)
	json.NewEncoder(w).Encode(resp)
}

// createTask 创建新任务
func createTask(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		resp := Response{Code: 405, Message: "方法不允许"}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(resp)
		return
	}

	var task Task
	if err := json.NewDecoder(r.Body).Decode(&task); err != nil {
		resp := Response{Code: 400, Message: "无效的请求体"}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(resp)
		return
	}

	task.ID = len(tasks) + 1
	tasks = append(tasks, task)

	resp := Response{
		Code:    201,
		Message: "创建成功",
		Data:    task,
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(resp)
}

// handleTasks 路由处理器
func handleTasks(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		if r.URL.Path == "/api/tasks" {
			getTasks(w, r)
		} else {
			getTaskByID(w, r)
		}
	} else if r.Method == http.MethodPost {
		createTask(w, r)
	} else {
		resp := Response{Code: 405, Message: "方法不允许"}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(resp)
	}
}

func main() {
	http.HandleFunc("/api/tasks", handleTasks)
	http.HandleFunc("/api/tasks/", handleTasks)

	fmt.Println("REST 服务启动于 http://localhost:8080")
	fmt.Println("GET  /api/tasks       - 获取所有任务")
	fmt.Println("GET  /api/tasks/{id}  - 获取单个任务")
	fmt.Println("POST /api/tasks       - 创建任务")
	http.ListenAndServe(":8080", nil)
}
