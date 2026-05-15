package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
	"gorm.io/gorm"

	"web-demo/enterprise/config"
	"web-demo/enterprise/internal/cache"
	"web-demo/enterprise/internal/database"
	"web-demo/enterprise/internal/repository"
	"web-demo/enterprise/internal/router"
	"web-demo/enterprise/pkg/logger"
)

var log zerolog.Logger

func main() {
	// 加载配置
	cfgPath := "config.yaml"
	if v := os.Getenv("CONFIG_PATH"); v != "" {
		cfgPath = v
	}
	cfg, err := config.Load(cfgPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "加载配置失败: %v\n", err)
		os.Exit(1)
	}

	// 初始化日志
	log = logger.New(cfg.Log)
	log.Info().Msg("=== 系统启动 ===")

	// 设置 Gin 模式
	gin.SetMode(cfg.Server.Mode)

	// 初始化缓存系统
	log.Info().Msg("初始化缓存系统...")
	c := cache.New(cfg, log)
	defer c.Close()

	// 初始化数据库
	log.Info().Msg("初始化数据库...")
	db := database.New(cfg, log)

	// 自动迁移
	log.Info().Msg("执行数据库迁移...")
	if err := repository.NewTaskRepo(db).AutoMigrate(); err != nil {
		log.Fatal().Err(err).Msg("数据库迁移失败")
	}

	// 初始化示例数据
	initSeedData(db)

	// 设置路由
	log.Info().Msg("配置路由...")
	r := router.Setup(cfg, db, c, log)

	// 创建 HTTP 服务器
	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Server.Port),
		Handler:      r,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		IdleTimeout:  cfg.Server.IdleTimeout,
	}

	// 打印启动信息
	printBanner(cfg.Server.Port)

	// 启动服务器的 goroutine
	go func() {
		log.Info().Str("addr", server.Addr).Msg("HTTP 服务器启动")
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatal().Err(err).Msg("服务启动失败")
		}
	}()

	// 等待中断信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info().Msg("正在关闭服务器...")

	// 优雅关闭（最多等待 30 秒）
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatal().Err(err).Msg("服务器强制关闭")
	}

	log.Info().Msg("服务器已安全关闭")
}

// initSeedData 初始化示例数据（仅首次启动时）
func initSeedData(db *gorm.DB) {
	repo := repository.NewTaskRepo(db)
	count, err := repo.Count()
	if err != nil {
		log.Warn().Err(err).Msg("查询数据量失败")
		return
	}

	if count == 0 {
		log.Info().Msg("初始化示例数据...")
		db.Exec("INSERT INTO tasks (title, done) VALUES ('学习 Go', false), ('构建 REST API', true)")
		log.Info().Msg("示例数据创建成功")
	}
}

// printBanner 打印启动横幅
func printBanner(port int) {
	fmt.Println("╔════════════════════════════════════════════╗")
	fmt.Println("║      Gin REST API + 2级缓存系统启动         ║")
	fmt.Println("╚════════════════════════════════════════════╝")
	fmt.Printf("📍 服务地址: http://localhost:%d\n", port)
	fmt.Println("")
	fmt.Println("📚 API 端点:")
	fmt.Println("  GET    /health           - 健康检查")
	fmt.Println("  GET    /health/liveness  - 存活检查")
	fmt.Println("  GET    /health/readiness - 就绪检查")
	fmt.Println("  GET    /api/tasks        - 获取所有任务")
	fmt.Println("  GET    /api/tasks/count  - 统计任务总数")
	fmt.Println("  GET    /api/tasks/{id}   - 获取单个任务")
	fmt.Println("  POST   /api/tasks        - 创建任务")
	fmt.Println("  PUT    /api/tasks/{id}   - 更新任务")
	fmt.Println("  DELETE /api/tasks/{id}   - 删除任务")
	fmt.Println("")
	fmt.Println("🧹 缓存管理:")
	fmt.Println("  DELETE /api/cache        - 清除所有缓存")
	fmt.Println("  DELETE /api/cache/{id}   - 清除任务缓存")
	fmt.Println("")
	fmt.Println("📖 API 文档:")
	fmt.Println("  GET    /swagger/*any     - Swagger UI")
	fmt.Println("")
	fmt.Println("按 Ctrl+C 优雅关闭服务...")
	fmt.Println("")
}
