package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	apperrors "web-demo/enterprise/errors"
)

// RateLimiter 简单的令牌桶限流器
type RateLimiter struct {
	mu       sync.Mutex
	requests map[string]*bucket
	rate     float64 // 每秒填充的令牌数
	burst    int     // 桶容量
}

type bucket struct {
	tokens    float64
	lastCheck time.Time
}

// NewRateLimiter 创建限流器
func NewRateLimiter(rate float64, burst int) *RateLimiter {
	rl := &RateLimiter{
		requests: make(map[string]*bucket),
		rate:     rate,
		burst:    burst,
	}

	// 定期清理过期记录
	go rl.cleanup()
	return rl
}

func (rl *RateLimiter) cleanup() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()
	for range ticker.C {
		rl.mu.Lock()
		for ip, b := range rl.requests {
			if time.Since(b.lastCheck) > 10*time.Minute {
				delete(rl.requests, ip)
			}
		}
		rl.mu.Unlock()
	}
}

// Allow 检查是否允许请求
func (rl *RateLimiter) Allow(ip string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	b, exists := rl.requests[ip]
	if !exists {
		b = &bucket{
			tokens:    float64(rl.burst),
			lastCheck: time.Now(),
		}
		rl.requests[ip] = b
	}

	// 计算新增令牌
	now := time.Now()
	elapsed := now.Sub(b.lastCheck).Seconds()
	b.tokens += elapsed * rl.rate
	if b.tokens > float64(rl.burst) {
		b.tokens = float64(rl.burst)
	}
	b.lastCheck = now

	if b.tokens >= 1 {
		b.tokens--
		return true
	}
	return false
}

// RateLimit 限流中间件
func RateLimit(rate float64, burst int) gin.HandlerFunc {
	limiter := NewRateLimiter(rate, burst)

	return func(c *gin.Context) {
		ip := c.ClientIP()
		if !limiter.Allow(ip) {
			c.PureJSON(http.StatusTooManyRequests, apperrors.ErrRateLimited)
			c.Abort()
			return
		}
		c.Next()
	}
}
