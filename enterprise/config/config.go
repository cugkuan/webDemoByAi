package config

import (
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

// Config 应用配置
type Config struct {
	Server   ServerConfig   `yaml:"server"`
	Database DatabaseConfig `yaml:"database"`
	Redis    RedisConfig    `yaml:"redis"`
	Cache    CacheConfig    `yaml:"cache"`
	Log      LogConfig      `yaml:"log"`
	Rate     RateConfig     `yaml:"rate"`
}

// ServerConfig 服务器配置
type ServerConfig struct {
	Port         int           `yaml:"port"`
	ReadTimeout  time.Duration `yaml:"read_timeout"`
	WriteTimeout time.Duration `yaml:"write_timeout"`
	IdleTimeout  time.Duration `yaml:"idle_timeout"`
	Mode         string        `yaml:"mode"`
}

// DatabaseConfig 数据库配置
type DatabaseConfig struct {
	DSN             string        `yaml:"dsn"`
	MaxOpenConns    int           `yaml:"max_open_conns"`
	MaxIdleConns    int           `yaml:"max_idle_conns"`
	ConnMaxLifetime time.Duration `yaml:"conn_max_lifetime"`
}

// RedisConfig Redis 配置
type RedisConfig struct {
	Addr     string `yaml:"addr"`
	Password string `yaml:"password"`
	DB       int    `yaml:"db"`
}

// CacheConfig 缓存配置
type CacheConfig struct {
	L1TTL time.Duration `yaml:"l1_ttl"`
	L2TTL time.Duration `yaml:"l2_ttl"`
}

// LogConfig 日志配置
type LogConfig struct {
	Level  string `yaml:"level"`
	Format string `yaml:"format"` // json 或 text
}

// RateConfig 限流配置
type RateConfig struct {
	Enabled  bool    `yaml:"enabled"`
	Requests int     `yaml:"requests"`
	PerSec   float64 `yaml:"per_sec"`
}

// DefaultConfig 返回默认配置
func DefaultConfig() *Config {
	return &Config{
		Server: ServerConfig{
			Port:         8080,
			ReadTimeout:  15 * time.Second,
			WriteTimeout: 15 * time.Second,
			IdleTimeout:  60 * time.Second,
			Mode:         "debug",
		},
		Database: DatabaseConfig{
			DSN:             "root@tcp(127.0.0.1:3306)/task_db?charset=utf8mb4&parseTime=True&loc=Local",
			MaxOpenConns:    25,
			MaxIdleConns:    10,
			ConnMaxLifetime: 5 * time.Minute,
		},
		Redis: RedisConfig{
			Addr: "localhost:6379",
			DB:   0,
		},
		Cache: CacheConfig{
			L1TTL: 30 * time.Second,
			L2TTL: 5 * time.Minute,
		},
		Log: LogConfig{
			Level:  "info",
			Format: "text",
		},
		Rate: RateConfig{
			Enabled:  true,
			Requests: 100,
			PerSec:   1.0,
		},
	}
}

// Load 从文件加载配置，支持环境变量覆盖
func Load(path string) (*Config, error) {
	cfg := DefaultConfig()

	// 尝试从 YAML 文件加载
	if path != "" {
		data, err := os.ReadFile(path)
		if err == nil {
			if err := yaml.Unmarshal(data, cfg); err != nil {
				return nil, err
			}
		}
	}

	// 环境变量覆盖
	if v := os.Getenv("SERVER_PORT"); v != "" {
		cfg.Server.Port = parseInt(v, cfg.Server.Port)
	}
	if v := os.Getenv("MYSQL_DSN"); v != "" {
		cfg.Database.DSN = v
	}
	if v := os.Getenv("REDIS_ADDR"); v != "" {
		cfg.Redis.Addr = v
	}
	if v := os.Getenv("LOG_LEVEL"); v != "" {
		cfg.Log.Level = v
	}
	if v := os.Getenv("GIN_MODE"); v != "" {
		cfg.Server.Mode = v
	}

	return cfg, nil
}

func parseInt(s string, defaultVal int) int {
	var v int
	for _, c := range s {
		if c < '0' || c > '9' {
			return defaultVal
		}
		v = v*10 + int(c-'0')
	}
	return v
}
