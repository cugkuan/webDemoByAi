package repository

import (
	"gorm.io/gorm"

	"web-demo/enterprise/internal/model"
)

// UserRepo 用户数据仓库
type UserRepo struct {
	db *gorm.DB
}

// NewUserRepo 创建用户仓库
func NewUserRepo(db *gorm.DB) *UserRepo {
	return &UserRepo{db: db}
}

// Create 创建用户
func (r *UserRepo) Create(user *model.User) error {
	return r.db.Create(user).Error
}

// FindByUsername 根据用户名查找用户
func (r *UserRepo) FindByUsername(username string) (*model.User, error) {
	var user model.User
	if err := r.db.Where("username = ?", username).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

// FindByID 根据 ID 查找用户
func (r *UserRepo) FindByID(id uint) (*model.User, error) {
	var user model.User
	if err := r.db.First(&user, id).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

// AutoMigrate 自动迁移
func (r *UserRepo) AutoMigrate() error {
	return r.db.AutoMigrate(&model.User{})
}
