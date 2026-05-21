package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"

	apperrors "web-demo/enterprise/errors"
	"web-demo/enterprise/internal/service"
)

// AuthRequired JWT 认证中间件（验证签名 + Redis 白名单）
func AuthRequired(userSvc *service.UserService, logger zerolog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			logger.Warn().Str("path", c.Request.URL.Path).Msg("缺少 Authorization 头")
			c.PureJSON(http.StatusUnauthorized, apperrors.ErrMissingToken)
			c.Abort()
			return
		}

		// 解析 Bearer Token
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			logger.Warn().Str("path", c.Request.URL.Path).Msg("Authorization 格式错误")
			c.PureJSON(http.StatusUnauthorized, apperrors.ErrInvalidToken)
			c.Abort()
			return
		}

		tokenString := parts[1]

		// 验证 JWT 签名 + Redis 白名单
		userID, err := userSvc.ValidateTokenWithRedis(tokenString)
		if err != nil {
			logger.Warn().Str("path", c.Request.URL.Path).Err(err).Msg("Token 验证失败")
			c.PureJSON(http.StatusUnauthorized, apperrors.ErrUnauthorized)
			c.Abort()
			return
		}

		// 将用户 ID 设置到 context 中
		c.Set("user_id", userID)
		c.Next()
	}
}
