package database

import (
	"os"

	"github.com/rs/zerolog"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"

	"web-demo/enterprise/config"
)

// New 根据配置创建数据库连接，并配置连接池
func New(cfg *config.Config, log zerolog.Logger) *gorm.DB {
	dsn := cfg.Database.DSN
	if dsn == "" {
		dsn = os.Getenv("MYSQL_DSN")
	}
	if dsn == "" {
		dsn = "root@tcp(127.0.0.1:3306)/task_db?charset=utf8mb4&parseTime=True&loc=Local"
		log.Warn().Msg("Using default DSN (no password); set MYSQL_DSN to override")
	}

	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal().Err(err).Msg("数据库连接失败")
	}

	sqlDB, err := db.DB()
	if err != nil {
		log.Fatal().Err(err).Msg("获取数据库实例失败")
	}
	sqlDB.SetMaxOpenConns(cfg.Database.MaxOpenConns)
	sqlDB.SetMaxIdleConns(cfg.Database.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(cfg.Database.ConnMaxLifetime)

	log.Info().
		Int("max_open", cfg.Database.MaxOpenConns).
		Int("max_idle", cfg.Database.MaxIdleConns).
		Dur("max_lifetime", cfg.Database.ConnMaxLifetime).
		Msg("数据库连接池配置完成")

	return db
}
