package service

import (
	"context"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"

	"web-demo/enterprise/config"
	apperrors "web-demo/enterprise/errors"
	"web-demo/enterprise/internal/cache"
	"web-demo/enterprise/internal/model"
	"web-demo/enterprise/internal/repository"
)

// TokenPrefix Redis Token 白名单前缀
const TokenPrefix = "token:"

// UserService 用户业务逻辑
type UserService struct {
	repo *repository.UserRepo
	cfg  *config.Config
	c    *cache.Cache
	log  zerolog.Logger
}

// NewUserService 创建用户服务
func NewUserService(repo *repository.UserRepo, cfg *config.Config, c *cache.Cache, log zerolog.Logger) *UserService {
	return &UserService{
		repo: repo,
		cfg:  cfg,
		c:    c,
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

	// 生成 JWT Token 并保存到 Redis 白名单
	token, jti, err := s.generateToken(user)
	if err != nil {
		s.log.Error().Err(err).Msg("生成 Token 失败")
		return nil, apperrors.ErrInternalServer
	}
	if err := s.saveTokenToRedis(jti, user.ID); err != nil {
		s.log.Error().Err(err).Msg("保存 Token 到 Redis 失败")
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

	// 生成 JWT Token 并保存到 Redis 白名单
	token, jti, err := s.generateToken(user)
	if err != nil {
		s.log.Error().Err(err).Msg("生成 Token 失败")
		return nil, apperrors.ErrInternalServer
	}
	if err := s.saveTokenToRedis(jti, user.ID); err != nil {
		s.log.Error().Err(err).Msg("保存 Token 到 Redis 失败")
		return nil, apperrors.ErrInternalServer
	}

	s.log.Debug().Uint("id", user.ID).Str("username", user.Username).Msg("SERVICE: 用户登录成功")
	return &model.LoginResponse{
		Token: token,
		User:  *user,
	}, nil
}

// Logout 退出登录（从 Redis 白名单中删除 token）
func (s *UserService) Logout(tokenString string) error {
	// 解析 token 获取 jti
	jti, err := s.extractJTI(tokenString)
	if err != nil {
		s.log.Warn().Err(err).Msg("退出登录：解析 Token 失败")
		return apperrors.ErrInvalidToken
	}

	// 从 Redis 中删除 token
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	key := TokenPrefix + jti
	if err := s.c.DeleteL2(ctx, key); err != nil {
		s.log.Error().Err(err).Str("jti", jti).Msg("退出登录：删除 Redis Token 失败")
		return apperrors.ErrInternalServer
	}

	s.log.Debug().Str("jti", jti).Msg("SERVICE: 退出登录成功")
	return nil
}

// ValidateToken 验证 JWT Token 签名并返回用户 ID（仅验证签名，不检查 Redis）
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

// ValidateTokenWithRedis 验证 JWT Token（签名 + Redis 白名单检查）
func (s *UserService) ValidateTokenWithRedis(tokenString string) (uint, error) {
	// 1. 验证 JWT 签名
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

	// 2. 提取 jti
	jti, ok := claims["jti"].(string)
	if !ok || jti == "" {
		s.log.Warn().Msg("Token 缺少 jti 字段")
		return 0, apperrors.ErrInvalidToken
	}

	// 3. 检查 Redis 白名单
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	key := TokenPrefix + jti
	var userID uint
	found, err := s.c.GetL2(ctx, key, &userID)
	if err != nil {
		s.log.Error().Err(err).Str("jti", jti).Msg("Redis 查询 Token 失败")
		return 0, apperrors.ErrInternalServer
	}
	if !found {
		s.log.Warn().Str("jti", jti).Msg("Token 不在 Redis 白名单中（可能已过期或已注销）")
		return 0, apperrors.ErrUnauthorized
	}

	return userID, nil
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

// generateToken 生成 JWT Token，返回 (tokenString, jti, error)
func (s *UserService) generateToken(user *model.User) (string, string, error) {
	jti := uuid.New().String()
	claims := jwt.MapClaims{
		"jti":      jti,
		"user_id":  user.ID,
		"username": user.Username,
		"iss":      s.cfg.JWT.Issuer,
		"iat":      time.Now().Unix(),
		"exp":      time.Now().Add(s.cfg.JWT.ExpireTime).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(s.cfg.JWT.Secret))
	if err != nil {
		return "", "", err
	}
	return tokenString, jti, nil
}

// saveTokenToRedis 将 token 的 jti 保存到 Redis 白名单
func (s *UserService) saveTokenToRedis(jti string, userID uint) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	key := TokenPrefix + jti
	// 使用与 JWT 相同的过期时间
	return s.c.SetL2WithTTL(ctx, key, userID, s.cfg.JWT.ExpireTime)
}

// extractJTI 从 token 字符串中提取 jti
func (s *UserService) extractJTI(tokenString string) (string, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(s.cfg.JWT.Secret), nil
	})
	if err != nil {
		return "", err
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		return "", fmt.Errorf("invalid token")
	}

	jti, ok := claims["jti"].(string)
	if !ok || jti == "" {
		return "", fmt.Errorf("token missing jti")
	}

	return jti, nil
}
