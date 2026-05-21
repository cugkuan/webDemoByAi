package service

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/rs/zerolog"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"web-demo/enterprise/config"
	apperrors "web-demo/enterprise/errors"
	"web-demo/enterprise/internal/model"
	"web-demo/enterprise/internal/repository"
)

// UserService 用户业务逻辑
type UserService struct {
	repo *repository.UserRepo
	cfg  *config.Config
	log  zerolog.Logger
}

// NewUserService 创建用户服务
func NewUserService(repo *repository.UserRepo, cfg *config.Config, log zerolog.Logger) *UserService {
	return &UserService{
		repo: repo,
		cfg:  cfg,
		log:  log,
	}
}

// Register 用户注册
func (s *UserService) Register(req *model.RegisterRequest) (*model.LoginResponse, error) {
	s.log.Debug().Str("username", req.Username).Msg("SERVICE: 用户注册")

	// 检查用户名是否已存在
	existing, err := s.repo.FindByUsername(req.Username)
	if err != nil && err != gorm.ErrRecordNotFound {
		s.log.Error().Err(err).Msg("查询用户失败")
		return nil, apperrors.ErrInternalServer
	}
	if existing != nil {
		s.log.Warn().Str("username", req.Username).Msg("用户名已存在")
		return nil, apperrors.ErrUsernameExists
	}

	// 密码加密
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		s.log.Error().Err(err).Msg("密码加密失败")
		return nil, apperrors.ErrInternalServer
	}

	// 创建用户
	user := &model.User{
		Username: req.Username,
		Password: string(hashedPassword),
	}
	if err := s.repo.Create(user); err != nil {
		s.log.Error().Err(err).Msg("创建用户失败")
		return nil, apperrors.ErrInternalServer
	}

	// 生成 JWT Token
	token, err := s.generateToken(user)
	if err != nil {
		s.log.Error().Err(err).Msg("生成 Token 失败")
		return nil, apperrors.ErrInternalServer
	}

	s.log.Debug().Uint("id", user.ID).Str("username", user.Username).Msg("SERVICE: 用户注册成功")
	return &model.LoginResponse{
		Token: token,
		User:  *user,
	}, nil
}

// Login 用户登录
func (s *UserService) Login(req *model.LoginRequest) (*model.LoginResponse, error) {
	s.log.Debug().Str("username", req.Username).Msg("SERVICE: 用户登录")

	// 查找用户
	user, err := s.repo.FindByUsername(req.Username)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			s.log.Warn().Str("username", req.Username).Msg("用户不存在")
			return nil, apperrors.ErrInvalidCredentials
		}
		s.log.Error().Err(err).Msg("查询用户失败")
		return nil, apperrors.ErrInternalServer
	}

	// 验证密码
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		s.log.Warn().Str("username", req.Username).Msg("密码错误")
		return nil, apperrors.ErrInvalidCredentials
	}

	// 生成 JWT Token
	token, err := s.generateToken(user)
	if err != nil {
		s.log.Error().Err(err).Msg("生成 Token 失败")
		return nil, apperrors.ErrInternalServer
	}

	s.log.Debug().Uint("id", user.ID).Str("username", user.Username).Msg("SERVICE: 用户登录成功")
	return &model.LoginResponse{
		Token: token,
		User:  *user,
	}, nil
}

// ValidateToken 验证 JWT Token 并返回用户 ID
func (s *UserService) ValidateToken(tokenString string) (uint, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(s.cfg.JWT.Secret), nil
	})

	if err != nil {
		return 0, apperrors.ErrUnauthorized
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return 0, apperrors.ErrUnauthorized
	}

	// 从 claims 中提取用户 ID
	userIDFloat, ok := claims["user_id"].(float64)
	if !ok {
		return 0, apperrors.ErrUnauthorized
	}

	return uint(userIDFloat), nil
}

// GetUserByID 根据 ID 获取用户
func (s *UserService) GetUserByID(id uint) (*model.User, error) {
	user, err := s.repo.FindByID(id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, apperrors.ErrNotFound
		}
		return nil, err
	}
	return user, nil
}

// generateToken 生成 JWT Token
func (s *UserService) generateToken(user *model.User) (string, error) {
	claims := jwt.MapClaims{
		"user_id":  user.ID,
		"username": user.Username,
		"iss":      s.cfg.JWT.Issuer,
		"iat":      time.Now().Unix(),
		"exp":      time.Now().Add(s.cfg.JWT.ExpireTime).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.cfg.JWT.Secret))
}
