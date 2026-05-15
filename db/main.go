package main

import (
	"fmt"
	"net/http"
)

func main() {
	// 初始化数据库
	InitDatabase()

	// 注册路由
	http.HandleFunc("/api/tasks", HandleTasks)
	http.HandleFunc("/api/tasks/", HandleTasks)

	// 启动服务
	fmt.Println("REST 服务启动于 http://localhost:8080 (使用 MySQL + GORM)")
	fmt.Println("GET    /api/tasks       - 获取所有任务")
	fmt.Println("GET    /api/tasks/{id}  - 获取单个任务")
	fmt.Println("POST   /api/tasks       - 创建任务")
	fmt.Println("PUT    /api/tasks/{id}  - 更新任务")
	fmt.Println("DELETE /api/tasks/{id}  - 删除任务")
	http.ListenAndServe(":8080", nil)
}
