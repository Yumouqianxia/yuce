package middleware

import (
	"net/http"
	"sync"
	"time"

	"backend-go/pkg/response"
	"github.com/gin-gonic/gin"
)

// RateLimiter 速率限制器
type RateLimiter struct {
	visitors map[string]*Visitor
	mu       sync.RWMutex
	rate     int           // 每个时间窗口允许的请求数
	window   time.Duration // 时间窗口
}

// Visitor 访问者信息
type Visitor struct {
	count     int
	lastReset time.Time
}

// NewRateLimiter 创建速率限制器
func NewRateLimiter(rate int, window time.Duration) *RateLimiter {
	rl := &RateLimiter{
		visitors: make(map[string]*Visitor),
		rate:     rate,
		window:   window,
	}

	// 启动清理协程
	go rl.cleanup()

	return rl
}

// Allow 检查是否允许请求
func (rl *RateLimiter) Allow(ip string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	visitor, exists := rl.visitors[ip]

	if !exists {
		rl.visitors[ip] = &Visitor{
			count:     1,
			lastReset: now,
		}
		return true
	}

	// 检查是否需要重置计数器
	if now.Sub(visitor.lastReset) > rl.window {
		visitor.count = 1
		visitor.lastReset = now
		return true
	}

	// 检查是否超过限制
	if visitor.count >= rl.rate {
		return false
	}

	visitor.count++
	return true
}

// cleanup 清理过期的访问者记录
func (rl *RateLimiter) cleanup() {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			rl.mu.Lock()
			now := time.Now()
			for ip, visitor := range rl.visitors {
				if now.Sub(visitor.lastReset) > rl.window*2 {
					delete(rl.visitors, ip)
				}
			}
			rl.mu.Unlock()
		}
	}
}

// RateLimit 速率限制中间件
func RateLimit(rate int, window time.Duration) gin.HandlerFunc {
	limiter := NewRateLimiter(rate, window)

	return func(c *gin.Context) {
		// 获取客户端IP
		ip := c.ClientIP()

		// 检查是否允许请求
		if !limiter.Allow(ip) {
			response.Error(c, http.StatusTooManyRequests, "Rate limit exceeded", "Too many requests, please try again later")
			c.Abort()
			return
		}

		c.Next()
	}
}
