package main

import (
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

// GetAllTasks 获取所有任务
func GetAllTasks() ([]Task, error) {
	var tasks []Task
	if err := DB.Order("id DESC").Find(&tasks).Error; err != nil {
		return nil, err
	}
	return tasks, nil
}

// GetTaskByID 根据 ID 获取任务
func GetTaskByID(id uint) (*Task, error) {
	var task Task
	if err := DB.First(&task, id).Error; err != nil {
		return nil, err
	}
	return &task, nil
}

// CreateTask 创建任务
func CreateTask(title string, done bool) (*Task, error) {
	task := Task{Title: title, Done: done}
	if err := DB.Create(&task).Error; err != nil {
		return nil, err
	}
	return &task, nil
}

// UpdateTask 更新任务
func UpdateTask(id uint, title string, done bool) (*Task, error) {
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
	
	return &task, nil
}

// DeleteTask 删除任务
func DeleteTask(id uint) error {
	return DB.Delete(&Task{}, id).Error
}
