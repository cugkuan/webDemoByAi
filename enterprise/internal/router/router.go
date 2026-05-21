package router

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"gorm.io/gorm"

	"web-demo/enterprise/config"
	_ "web-demo/enterprise/docs"
	"web-demo/enterprise/internal/cache"
	"web-demo/enterprise/internal/handler"
	"web-demo/enterprise/internal/repository"
	"web-demo/enterprise/internal/service"
	"web-demo/enterprise/middleware"
)

// Setup 设置路由，返回 *gin.Engine
func Setup(cfg *config.Config, db *gorm.DB, c *cache.Cache, log zerolog.Logger) *gin.Engine {
	router := gin.New()

	// 全局中间件

	router.Use(gin.Recovery())
	router.Use(middleware.RequestID())
	router.Use(middleware.Logger(log))
	router.Use(middleware.ErrorHandler(log))

	// CORS 中间件（允许 Authorization 头）
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "X-Request-ID", "Authorization"},
		ExposeHeaders:    []string{"X-Request-ID"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// 限流中间件（如果启用）
	if cfg.Rate.Enabled {
		router.Use(middleware.RateLimit(cfg.Rate.PerSec, cfg.Rate.Requests))
	}

	// 依赖注入
	taskRepo := repository.NewTaskRepo(db)
	taskSvc := service.NewTaskService(taskRepo, c, log)
	taskH := handler.NewTaskHandler(taskSvc, c, log)

	// 用户相关依赖注入
	userRepo := repository.NewUserRepo(db)
	userSvc := service.NewUserService(userRepo, cfg, c, log)
	authH := handler.NewAuthHandler(userSvc, log)

	// Swagger API 文档
	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerfiles.Handler))

	// 健康检查
	router.GET("/health", healthCheck(db, c))
	router.GET("/health/liveness", livenessCheck)
	router.GET("/health/readiness", readinessCheck(db))

	// ============================================================
	// 认证路由（无需登录）
	// ============================================================
	authGroup := router.Group("/api/auth")
	{
		authGroup.POST("/register", authH.Register)
		authGroup.POST("/login", authH.Login)
	}

	// ============================================================
	// 需要认证的路由
	// ============================================================
	protected := router.Group("/api")
	protected.Use(middleware.AuthRequired(userSvc, log))
	{
		// 退出登录
		protected.POST("/auth/logout", authH.Logout)

		// 用户信息
		protected.GET("/profile", authH.GetProfile)

		// 任务相关路由
		taskGroup := protected.Group("/tasks")
		{
			taskGroup.GET("", taskH.GetTasks)
			taskGroup.GET("/count", taskH.CountTasks)
			taskGroup.GET("/:id", taskH.GetTask)
			taskGroup.POST("", taskH.CreateTask)
			taskGroup.PUT("/:id", taskH.UpdateTask)
			taskGroup.DELETE("/:id", taskH.DeleteTask)
		}

		// 缓存管理
		cacheGroup := protected.Group("/cache")
		{
			cacheGroup.DELETE("", taskH.ClearAllCache)
			cacheGroup.DELETE("/:id", taskH.ClearTaskCache)
		}
	}

	// 系统信息（无需认证）
	router.GET("/sys/info", sysInfo(c))
	router.GET("/sys/stats", sysStats)

	return router
}

// ============================================================
// 健康检查
// ============================================================

func healthCheck(db *gorm.DB, c *cache.Cache) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		status := map[string]interface{}{
			"status": "ok",
			"db":     checkDBHealth(db),
			"redis":  checkRedisHealth(c),
		}
		ctx.PureJSON(http.StatusOK, gin.H{
			"code":    200,
			"message": "成功",
			"data":    status,
		})
	}
}

func livenessCheck(ctx *gin.Context) {
	ctx.PureJSON(http.StatusOK, gin.H{"status": "alive"})
}

func readinessCheck(db *gorm.DB) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		dbStatus := checkDBHealth(db)
		if dbStatus != "ok" {
			ctx.PureJSON(http.StatusServiceUnavailable, gin.H{
				"status": "not_ready",
				"reason": "database " + dbStatus,
			})
			return
		}
		ctx.PureJSON(http.StatusOK, gin.H{"status": "ready"})
	}
}

// ============================================================
// 系统信息
// ============================================================

func sysInfo(c *cache.Cache) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		info := map[string]interface{}{
			"version": "1.0.0",
			"service": "Task API with 2-Level Cache",
			"cache": map[string]string{
				"L1_TTL": c.L1TTL().String(),
				"L2_TTL": c.L2TTL().String(),
			},
		}
		ctx.PureJSON(http.StatusOK, gin.H{
			"code":    200,
			"message": "成功",
			"data":    info,
		})
	}
}

func sysStats(ctx *gin.Context) {
	stats := map[string]interface{}{
		"uptime": "运行中",
		"cache": map[string]interface{}{
			"status": "运行中",
		},
		"database": map[string]interface{}{
			"status": "已连接",
		},
	}
	ctx.PureJSON(http.StatusOK, gin.H{
		"code":    200,
		"message": "成功",
		"data":    stats,
	})
}

// ============================================================
// 辅助函数
// ============================================================

func checkDBHealth(db *gorm.DB) string {
	if db == nil {
		return "disconnected"
	}
	sqlDB, err := db.DB()
	if err != nil {
		return "error"
	}
	if err := sqlDB.Ping(); err != nil {
		return "error"
	}
	return "ok"
}

func checkRedisHealth(c *cache.Cache) string {
	if c == nil {
		return "not_available"
	}
	// 尝试 Ping
	if err := c.Ping(context.Background()); err != nil {
		return "error"
	}
	return "ok"
}
