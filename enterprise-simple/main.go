package main

import (
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
)

func main() {
	// 初始化缓存系统
	InitCache()
	defer func() {
		if redisClient != nil {
			redisClient.Close()
		}
	}()
	
	// 初始化数据库
	InitDatabase()

	// 设置路由
	router := SetupRouter()

	// 创建 HTTP 服务器（带超时配置）
	server := &HTTPServer{
		Engine: router,
		Addr:   ":8080",
		// 读取请求超时（客户端发送请求的超时）
		ReadTimeout: 15 * time.Second,
		// 写入响应超时（服务器写响应的超时）
		WriteTimeout: 15 * time.Second,
		// 空闲连接超时（keep-alive 连接的超时）
		IdleTimeout: 60 * time.Second,
	}

	// 启动服务
	fmt.Println("╔════════════════════════════════════════════╗")
	fmt.Println("║      Gin REST API + 2级缓存系统启动         ║")
	fmt.Println("╚════════════════════════════════════════════╝")
	fmt.Println("")
	fmt.Println("📍 服务地址: http://localhost:8080")
	fmt.Println("")
	fmt.Println("📚 API 端点:")
	fmt.Println("  GET    /health           - 健康检查")
	fmt.Println("  GET    /api/tasks        - 获取所有任务")
	fmt.Println("  GET    /api/tasks/{id}   - 获取单个任务")
	fmt.Println("  POST   /api/tasks        - 创建任务")
	fmt.Println("  PUT    /api/tasks/{id}   - 更新任务")
	fmt.Println("  DELETE /api/tasks/{id}   - 删除任务")
	fmt.Println("")
	fmt.Println("🧹 缓存管理:")
	fmt.Println("  DELETE /api/cache        - 清除所有缓存")
	fmt.Println("  DELETE /api/cache/{id}   - 清除任务缓存")
	fmt.Println("")
	fmt.Println("ℹ️  系统信息:")
	fmt.Println("  GET    /sys/info         - 系统信息")
	fmt.Println("  GET    /sys/stats        - 系统统计")
	fmt.Println("")
	fmt.Println("⏱️  超时配置:")
	fmt.Println("  读超时: 15 秒 | 写超时: 15 秒 | 空闲超时: 60 秒")
	fmt.Println("")
	fmt.Println("🚀 Gin 框架模式:", gin.Mode())
	fmt.Println("")
	fmt.Println("按 Ctrl+C 优雅关闭服务...")
	fmt.Println("")

	if err := server.ListenAndServe(); err != nil {
		fmt.Printf("服务启动失败: %v\n", err)
	}
}
