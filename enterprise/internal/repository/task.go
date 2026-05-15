package repository

import (
	"gorm.io/gorm"

	"web-demo/enterprise/internal/model"
)

// TaskRepo 任务数据仓库
type TaskRepo struct {
	db *gorm.DB
}

// NewTaskRepo 创建任务仓库
func NewTaskRepo(db *gorm.DB) *TaskRepo {
	return &TaskRepo{db: db}
}

// FindAll 获取所有任务
func (r *TaskRepo) FindAll() ([]model.Task, error) {
	var tasks []model.Task
	if err := r.db.Order("id DESC").Find(&tasks).Error; err != nil {
		return nil, err
	}
	return tasks, nil
}

// FindByID 根据 ID 获取任务
func (r *TaskRepo) FindByID(id uint) (*model.Task, error) {
	var task model.Task
	if err := r.db.First(&task, id).Error; err != nil {
		return nil, err
	}
	return &task, nil
}

// Create 创建任务
func (r *TaskRepo) Create(task *model.Task) error {
	return r.db.Create(task).Error
}

// Update 更新任务
func (r *TaskRepo) Update(task *model.Task) error {
	return r.db.Save(task).Error
}

// Delete 删除任务
func (r *TaskRepo) Delete(id uint) error {
	return r.db.Delete(&model.Task{}, id).Error
}

// AutoMigrate 自动迁移
func (r *TaskRepo) AutoMigrate() error {
	return r.db.AutoMigrate(&model.Task{})
}

// Count 统计任务数量
func (r *TaskRepo) Count() (int64, error) {
	var count int64
	if err := r.db.Model(&model.Task{}).Count(&count).Error; err != nil {
		return 0, err
	}
	return count, nil
}
