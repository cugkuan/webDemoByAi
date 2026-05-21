package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"

	apperrors "web-demo/enterprise/errors"
	"web-demo/enterprise/internal/model"
	"web-demo/enterprise/internal/service"
	"web-demo/enterprise/pkg/response"
)

// AuthHandler 认证 HTTP Handler
type AuthHandler struct {
	svc *service.UserService
	log zerolog.Logger
}

// NewAuthHandler 创建认证 Handler
func NewAuthHandler(svc *service.UserService, log zerolog.Logger) *AuthHandler {
	return &AuthHandler{
		svc: svc,
		log: log,
	}
}

// Register 用户注册
func (h *AuthHandler) Register(c *gin.Context) {
	var req model.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.PureJSON(http.StatusBadRequest, apperrors.ErrBadRequest)
		return
	}

	// 参数校验
	if len(req.Username) < 3 || len(req.Username) > 50 {
		c.PureJSON(http.StatusBadRequest, apperrors.New(400, "用户名长度必须在 3-50 个字符之间"))
		return
	}
	if len(req.Password) < 6 || len(req.Password) > 100 {
		c.PureJSON(http.StatusBadRequest, apperrors.New(400, "密码长度必须在 6-100 个字符之间"))
		return
	}

	result, err := h.svc.Register(&req)
	if err != nil {
		// 处理已知错误
		if appErr, ok := err.(*apperrors.AppError); ok {
			c.PureJSON(apperrors.HTTPStatus(appErr.Code), appErr)
			return
		}
		c.Error(err)
		return
	}

	c.PureJSON(http.StatusCreated, response.Created(result))
}

// Login 用户登录
func (h *AuthHandler) Login(c *gin.Context) {
	var req model.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.PureJSON(http.StatusBadRequest, apperrors.ErrBadRequest)
		return
	}

	result, err := h.svc.Login(&req)
	if err != nil {
		// 处理已知错误
		if appErr, ok := err.(*apperrors.AppError); ok {
			c.PureJSON(apperrors.HTTPStatus(appErr.Code), appErr)
			return
		}
		c.Error(err)
		return
	}

	c.PureJSON(http.StatusOK, response.Success(result))
}

// GetProfile 获取当前用户信息（需要认证）
func (h *AuthHandler) GetProfile(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.PureJSON(http.StatusUnauthorized, apperrors.ErrUnauthorized)
		return
	}

	user, err := h.svc.GetUserByID(userID.(uint))
	if err != nil {
		if appErr, ok := err.(*apperrors.AppError); ok {
			c.PureJSON(apperrors.HTTPStatus(appErr.Code), appErr)
			return
		}
		c.Error(err)
		return
	}

	c.PureJSON(http.StatusOK, response.Success(user))
}
