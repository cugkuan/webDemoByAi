package logger

import (
	"io"
	"os"
	"time"

	"github.com/rs/zerolog"

	"web-demo/enterprise/config"
)

// New 根据配置创建 zerolog.Logger
// 支持同时输出到控制台和文件
func New(cfg config.LogConfig) zerolog.Logger {
	level, err := zerolog.ParseLevel(cfg.Level)
	if err != nil {
		level = zerolog.InfoLevel
	}

	// 构建输出目标：控制台 + 可选的日志文件
	var writers []io.Writer

	// 控制台输出
	if cfg.Format == "json" {
		writers = append(writers, os.Stdout)
	} else {
		writers = append(writers, zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.RFC3339})
	}

	// 日志文件输出（如果配置了 file_path）
	if cfg.FilePath != "" {
		// 确保目录存在
		if dir := cfg.FilePath[:len(cfg.FilePath)-len(cfg.FilePath)]; dir != "" {
			os.MkdirAll(dir, 0755)
		}

		file, err := os.OpenFile(cfg.FilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err == nil {
			writers = append(writers, file)
		}
	}

	// 多输出目标
	multiWriter := io.MultiWriter(writers...)

	return zerolog.New(multiWriter).
		Level(level).
		With().
		Timestamp().
		Caller().
		Logger()
}
