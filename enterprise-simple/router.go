package main

import (
	"context"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// SetupRouter 设置路由
func SetupRouter() *gin.Engine {
	// 在生产环境设置为 release 模式
	// gin.SetMode(gin.ReleaseMode)
	
	router := gin.Default()

	// 添加中间件
	router.Use(loggingMiddleware())
	router.Use(recoveryMiddleware())

	// 健康检查
	router.GET("/health", healthCheck)

	// 任务相关路由
	taskGroup := router.Group("/api/tasks")
	{
		// 获取所有任务
		taskGroup.GET("", getTasks)

		// 获取单个任务
		taskGroup.GET("/:id", getTask)

		// 创建任务
		taskGroup.POST("", createTask)

		// 更新任务
		taskGroup.PUT("/:id", updateTask)

		// 删除任务
		taskGroup.DELETE("/:id", deleteTask)
	}

	// 缓存管理
	cacheGroup := router.Group("/api/cache")
	{
		// 清除所有缓存
		cacheGroup.DELETE("", clearAllCache)

		// 清除指定任务缓存
		cacheGroup.DELETE("/:id", clearTaskCache)
	}

	// 系统信息
	router.GET("/sys/info", sysInfo)
	router.GET("/sys/stats", sysStats)

	return router
}

// ============================================================
// 中间件
// ============================================================

// loggingMiddleware 日志中间件
func loggingMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 继续处理请求
		c.Next()

		// 记录请求信息（这里可以集成真实的日志系统）
		statusCode := c.Writer.Status()
		latency := c.GetDuration("latency")
		
		// 后续添加实际日志系统
		_ = statusCode
		_ = latency
	}
}

// recoveryMiddleware 错误恢复中间件
func recoveryMiddleware() gin.HandlerFunc {
	return gin.Recovery()
}

// ============================================================
// Handler 函数
// ============================================================

// healthCheck 健康检查
// @Summary 健康检查
// @Description 检查服务和依赖的健康状态
// @Tags system
// @Accept json
// @Produce json
// @Success 200 {object} HealthStatus
// @Router /health [get]
func healthCheck(c *gin.Context) {
	status := map[string]interface{}{
		"status": "ok",
		"db":     checkDatabaseHealth(),
		"redis":  checkRedisHealth(),
	}
	c.JSON(http.StatusOK, status)
}

// getTasks 获取所有任务
// @Summary 获取所有任务
// @Description 返回任务列表，支持缓存
// @Tags tasks
// @Accept json
// @Produce json
// @Success 200 {object} Response
// @Router /api/tasks [get]
func getTasks(c *gin.Context) {
	tasks, err := GetAllTasks()
	if err != nil {
		c.JSON(http.StatusInternalServerError, Response{
			Code:    500,
			Message: "数据库错误",
		})
		return
	}

	c.JSON(http.StatusOK, Response{
		Code:    200,
		Message: "成功",
		Data:    tasks,
	})
}

// getTask 获取单个任务
// @Summary 获取单个任务
// @Description 根据 ID 获取任务详情
// @Tags tasks
// @Accept json
// @Produce json
// @Param id path int true "Task ID"
// @Success 200 {object} Response
// @Failure 404 {object} Response
// @Router /api/tasks/{id} [get]
func getTask(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, Response{
			Code:    400,
			Message: "无效的任务 ID",
		})
		return
	}

	task, err := GetTaskByID(uint(id))
	if err == gorm.ErrRecordNotFound {
		c.JSON(http.StatusNotFound, Response{
			Code:    404,
			Message: "任务不存在",
		})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, Response{
			Code:    500,
			Message: "数据库错误",
		})
		return
	}

	c.JSON(http.StatusOK, Response{
		Code:    200,
		Message: "成功",
		Data:    task,
	})
}

// createTask 创建任务
// @Summary 创建任务
// @Description 创建新任务
// @Tags tasks
// @Accept json
// @Produce json
// @Param task body CreateTaskRequest true "Task data"
// @Success 201 {object} Response
// @Failure 400 {object} Response
// @Router /api/tasks [post]
func createTask(c *gin.Context) {
	var input struct {
		Title string `json:"title"`
		Done  bool   `json:"done"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, Response{
			Code:    400,
			Message: "无效的请求体",
		})
		return
	}

	if input.Title == "" {
		c.JSON(http.StatusBadRequest, Response{
			Code:    400,
			Message: "标题不能为空",
		})
		return
	}

	task, err := CreateTask(input.Title, input.Done)
	if err != nil {
		c.JSON(http.StatusInternalServerError, Response{
			Code:    500,
			Message: "数据库错误",
		})
		return
	}

	c.JSON(http.StatusCreated, Response{
		Code:    201,
		Message: "创建成功",
		Data:    task,
	})
}

// updateTask 更新任务
// @Summary 更新任务
// @Description 更新任务信息
// @Tags tasks
// @Accept json
// @Produce json
// @Param id path int true "Task ID"
// @Param task body UpdateTaskRequest true "Task data"
// @Success 200 {object} Response
// @Failure 404 {object} Response
// @Router /api/tasks/{id} [put]
func updateTask(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, Response{
			Code:    400,
			Message: "无效的任务 ID",
		})
		return
	}

	var input struct {
		Title string `json:"title"`
		Done  bool   `json:"done"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, Response{
			Code:    400,
			Message: "无效的请求体",
		})
		return
	}

	task, err := UpdateTask(uint(id), input.Title, input.Done)
	if err == gorm.ErrRecordNotFound {
		c.JSON(http.StatusNotFound, Response{
			Code:    404,
			Message: "任务不存在",
		})
		return
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, Response{
			Code:    500,
			Message: "更新失败",
		})
		return
	}

	c.JSON(http.StatusOK, Response{
		Code:    200,
		Message: "更新成功",
		Data:    task,
	})
}

// deleteTask 删除任务
// @Summary 删除任务
// @Description 删除指定任务
// @Tags tasks
// @Accept json
// @Produce json
// @Param id path int true "Task ID"
// @Success 200 {object} Response
// @Failure 404 {object} Response
// @Router /api/tasks/{id} [delete]
func deleteTask(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, Response{
			Code:    400,
			Message: "无效的任务 ID",
		})
		return
	}

	if err := DeleteTask(uint(id)); err != nil {
		c.JSON(http.StatusInternalServerError, Response{
			Code:    500,
			Message: "删除失败",
		})
		return
	}

	c.JSON(http.StatusOK, Response{
		Code:    200,
		Message: "删除成功",
	})
}

// clearAllCache 清除所有缓存
// @Summary 清除所有缓存
// @Description 手动清除 L1 和 L2 缓存
// @Tags cache
// @Accept json
// @Produce json
// @Success 200 {object} Response
// @Router /api/cache [delete]
func clearAllCache(c *gin.Context) {
	invalidateAllTasksCache(context.Background())
	c.JSON(http.StatusOK, Response{
		Code:    200,
		Message: "缓存已清除",
	})
}

// clearTaskCache 清除指定任务缓存
// @Summary 清除任务缓存
// @Description 清除指定任务的缓存
// @Tags cache
// @Accept json
// @Produce json
// @Param id path int true "Task ID"
// @Success 200 {object} Response
// @Router /api/cache/{id} [delete]
func clearTaskCache(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, Response{
			Code:    400,
			Message: "无效的任务 ID",
		})
		return
	}

	invalidateTaskCache(context.Background(), uint(id))
	c.JSON(http.StatusOK, Response{
		Code:    200,
		Message: "任务缓存已清除",
	})
}

// sysInfo 系统信息
// @Summary 系统信息
// @Description 获取系统版本和配置信息
// @Tags system
// @Accept json
// @Produce json
// @Success 200 {object} Response
// @Router /sys/info [get]
func sysInfo(c *gin.Context) {
	info := map[string]interface{}{
		"version": "1.0.0",
		"service": "Task API with 2-Level Cache",
		"cache": map[string]string{
			"L1_TTL": "30s",
			"L2_TTL": "5m",
		},
		"timeout": map[string]string{
			"read":  "15s",
			"write": "15s",
			"idle":  "60s",
		},
	}

	c.JSON(http.StatusOK, Response{
		Code:    200,
		Message: "成功",
		Data:    info,
	})
}

// sysStats 系统统计
// @Summary 系统统计
// @Description 获取系统运行统计信息
// @Tags system
// @Accept json
// @Produce json
// @Success 200 {object} Response
// @Router /sys/stats [get]
func sysStats(c *gin.Context) {
	stats := map[string]interface{}{
		"uptime": "计算启动时间...",
		"cache": map[string]interface{}{
			"status": "运行中",
		},
		"database": map[string]interface{}{
			"status": "已连接",
		},
	}

	c.JSON(http.StatusOK, Response{
		Code:    200,
		Message: "成功",
		Data:    stats,
	})
}

// ============================================================
// 辅助函数
// ============================================================

// checkDatabaseHealth 检查数据库健康状态
func checkDatabaseHealth() string {
	if DB == nil {
		return "disconnected"
	}
	return "ok"
}

// checkRedisHealth 检查 Redis 健康状态
func checkRedisHealth() string {
	if redisClient == nil {
		return "not_available"
	}
	if err := redisClient.Ping(context.Background()).Err(); err != nil {
		return "error"
	}
	return "ok"
}

// HealthStatus 健康状态响应结构
type HealthStatus struct {
	Status string `json:"status"`
	DB     string `json:"db"`
	Redis  string `json:"redis"`
}

// CreateTaskRequest 创建任务请求
type CreateTaskRequest struct {
	Title string `json:"title" binding:"required"`
	Done  bool   `json:"done"`
}

// UpdateTaskRequest 更新任务请求
type UpdateTaskRequest struct {
	Title string `json:"title" binding:"required"`
	Done  bool   `json:"done"`
}
