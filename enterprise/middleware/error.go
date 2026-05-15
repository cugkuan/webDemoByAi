package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
	"gorm.io/gorm"

	apperrors "web-demo/enterprise/errors"
)

// ErrorHandler 统一错误处理中间件
func ErrorHandler(logger zerolog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		// 检查是否有错误
		if len(c.Errors) > 0 {
			err := c.Errors.Last().Err
			requestID, _ := c.Get("request_id")

			switch e := err.(type) {
			case *apperrors.AppError:
				logger.Warn().
					Str("request_id", toString(requestID)).
					Str("path", c.Request.URL.Path).
					Int("code", e.Code).
					Str("error", e.Error()).
					Msg("application error")
				c.JSON(apperrors.HTTPStatus(e.Code), e)
			default:
				// 处理 GORM 错误
				if err == gorm.ErrRecordNotFound {
					c.JSON(http.StatusNotFound, apperrors.ErrNotFound)
					return
				}

				logger.Error().
					Str("request_id", toString(requestID)).
					Str("path", c.Request.URL.Path).
					Str("error", err.Error()).
					Msg("internal server error")
				c.JSON(http.StatusInternalServerError, apperrors.ErrInternalServer)
			}

			// 阻止后续 handler 继续写入
			c.Abort()
		}
	}
}
