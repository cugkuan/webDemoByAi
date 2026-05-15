package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	apperrors "web-demo/enterprise/errors"
)

var validate = validator.New()

// ValidateRequest 验证请求体
func ValidateRequest(obj interface{}) gin.HandlerFunc {
	return func(c *gin.Context) {
		if err := c.ShouldBindJSON(obj); err != nil {
			// 处理验证错误
			if ve, ok := err.(validator.ValidationErrors); ok {
				msg := ve[0].Field() + " 验证失败: " + ve[0].Tag()
				c.JSON(http.StatusBadRequest, apperrors.New(400, msg))
				c.Abort()
				return
			}
			c.JSON(http.StatusBadRequest, apperrors.ErrBadRequest)
			c.Abort()
			return
		}

		// 额外验证
		if err := validate.Struct(obj); err != nil {
			if ve, ok := err.(validator.ValidationErrors); ok {
				msg := ve[0].Field() + " 验证失败: " + ve[0].Tag()
				c.JSON(http.StatusBadRequest, apperrors.New(400, msg))
				c.Abort()
				return
			}
		}

		c.Set("validated_body", obj)
		c.Next()
	}
}
