package main

import (
	"context"
	"fmt"
	"os"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var DB *gorm.DB

// InitDatabase 初始化数据库
func InitDatabase() {
	// MySQL 连接字符串
	// 格式: user:password@tcp(host:port)/database?charset=utf8mb4&parseTime=True&loc=Local
	dsn := os.Getenv("MYSQL_DSN")
	if dsn == "" {
		dsn = "root@tcp(127.0.0.1:3306)/task_db?charset=utf8mb4&parseTime=True&loc=Local"
		fmt.Println("Using default DSN (no password); set MYSQL_DSN to override")
	}

	var err error
	DB, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		panic("数据库连接失败: " + err.Error())
	}

	// 自动迁移创建表
	if err := DB.AutoMigrate(&Task{}); err != nil {
		panic("数据库表迁移失败: " + err.Error())
	}

	// 初始化数据
	var count int64
	DB.Model(&Task{}).Count(&count)
	if count == 0 {
		DB.Create(&Task{Title: "学习 Go", Done: false})
		DB.Create(&Task{Title: "构建 REST API", Done: true})
	}

	fmt.Println("MySQL 数据库连接成功")
}

// GetAllTasks 获取所有任务（带缓存）
func GetAllTasks() ([]Task, error) {
	ctx := context.Background()
	
	// L1 缓存检查
	if val, ok := getL1(AllTasksCacheKey); ok {
		return val.([]Task), nil
	}
	
	// L2 缓存检查
	var tasks []Task
	if found, err := getL2(ctx, AllTasksCacheKey, &tasks); found && err == nil {
		setL1(AllTasksCacheKey, tasks) // 回写到 L1
		return tasks, nil
	}
	
	// 从数据库查询
	if err := DB.Order("id DESC").Find(&tasks).Error; err != nil {
		return nil, err
	}
	
	// 写入缓存（L1 和 L2）
	setL1(AllTasksCacheKey, tasks)
	setL2(ctx, AllTasksCacheKey, tasks)
	
	return tasks, nil
}

// GetTaskByID 根据 ID 获取任务（带缓存）
func GetTaskByID(id uint) (*Task, error) {
	ctx := context.Background()
	key := fmt.Sprintf("%s%d", TaskCacheKeyPrefix, id)
	
	// L1 缓存检查
	if val, ok := getL1(key); ok {
		task := val.(*Task)
		return task, nil
	}
	
	// L2 缓存检查
	var task Task
	if found, err := getL2(ctx, key, &task); found && err == nil {
		setL1(key, &task) // 回写到 L1
		return &task, nil
	}
	
	// 从数据库查询
	if err := DB.First(&task, id).Error; err != nil {
		return nil, err
	}
	
	// 写入缓存（L1 和 L2）
	setL1(key, &task)
	setL2(ctx, key, &task)
	
	return &task, nil
}

// CreateTask 创建任务（清除缓存）
func CreateTask(title string, done bool) (*Task, error) {
	task := Task{Title: title, Done: done}
	if err := DB.Create(&task).Error; err != nil {
		return nil, err
	}
	
	// 清除任务列表缓存
	invalidateAllTasksCache(context.Background())
	
	return &task, nil
}

// UpdateTask 更新任务（清除缓存）
func UpdateTask(id uint, title string, done bool) (*Task, error) {
	ctx := context.Background()
	var task Task
	
	// 先检查任务是否存在
	if err := DB.First(&task, id).Error; err != nil {
		return nil, err
	}
	
	// 更新任务
	task.Title = title
	task.Done = done
	if err := DB.Save(&task).Error; err != nil {
		return nil, err
	}
	
	// 清除相关缓存
	invalidateTaskCache(ctx, id)
	
	return &task, nil
}

// DeleteTask 删除任务（清除缓存）
func DeleteTask(id uint) error {
	ctx := context.Background()
	if err := DB.Delete(&Task{}, id).Error; err != nil {
		return err
	}
	
	// 清除相关缓存
	invalidateTaskCache(ctx, id)
	
	return nil
}
