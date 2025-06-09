package middleware

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

type RateLimiter struct {
	visitors map[string]*Visitor
	mutex    sync.RWMutex
	rate     time.Duration
	limit    int
}

type Visitor struct {
	limiter  *TokenBucket
	lastSeen time.Time
}

type TokenBucket struct {
	tokens    int
	maxTokens int
	refillRate time.Duration
	lastRefill time.Time
	mutex     sync.Mutex
}

type RateLimitConfig struct {
	RequestsPerMinute int
	BurstSize         int
	KeyFunc           func(*gin.Context) string
	SkipSuccessful    bool
	SkipFailedAuth    bool
	Message           string
}

func NewTokenBucket(maxTokens int, refillRate time.Duration) *TokenBucket {
	return &TokenBucket{
		tokens:     maxTokens,
		maxTokens:  maxTokens,
		refillRate: refillRate,
		lastRefill: time.Now(),
	}
}

func (tb *TokenBucket) Allow() bool {
	tb.mutex.Lock()
	defer tb.mutex.Unlock()

	now := time.Now()
	elapsed := now.Sub(tb.lastRefill)
	
	// Refill tokens based on elapsed time
	tokensToAdd := int(elapsed / tb.refillRate)
	if tokensToAdd > 0 {
		tb.tokens += tokensToAdd
		if tb.tokens > tb.maxTokens {
			tb.tokens = tb.maxTokens
		}
		tb.lastRefill = now
	}

	if tb.tokens > 0 {
		tb.tokens--
		return true
	}

	return false
}

func NewRateLimiter(requestsPerMinute, burstSize int) *RateLimiter {
	return &RateLimiter{
		visitors: make(map[string]*Visitor),
		rate:     time.Minute / time.Duration(requestsPerMinute),
		limit:    burstSize,
	}
}

func (rl *RateLimiter) cleanupVisitors() {
	rl.mutex.Lock()
	defer rl.mutex.Unlock()

	for ip, visitor := range rl.visitors {
		if time.Since(visitor.lastSeen) > 3*time.Minute {
			delete(rl.visitors, ip)
		}
	}
}

func (rl *RateLimiter) getVisitor(key string) *Visitor {
	rl.mutex.Lock()
	defer rl.mutex.Unlock()

	visitor, exists := rl.visitors[key]
	if !exists {
		limiter := NewTokenBucket(rl.limit, rl.rate)
		visitor = &Visitor{
			limiter:  limiter,
			lastSeen: time.Now(),
		}
		rl.visitors[key] = visitor

		// Cleanup old visitors occasionally
		if len(rl.visitors)%100 == 0 {
			go rl.cleanupVisitors()
		}
	}

	visitor.lastSeen = time.Now()
	return visitor
}

func (rl *RateLimiter) Allow(key string) bool {
	visitor := rl.getVisitor(key)
	return visitor.limiter.Allow()
}

func DefaultKeyFunc(c *gin.Context) string {
	return c.ClientIP()
}

func UserKeyFunc(c *gin.Context) string {
	if userID, exists := c.Get("user_id"); exists {
		return fmt.Sprintf("user:%v", userID)
	}
	return c.ClientIP()
}

func RateLimit(config *RateLimitConfig) gin.HandlerFunc {
	if config == nil {
		config = &RateLimitConfig{
			RequestsPerMinute: 60,
			BurstSize:         10,
			KeyFunc:           DefaultKeyFunc,
			Message:           "Rate limit exceeded",
		}
	}

	if config.KeyFunc == nil {
		config.KeyFunc = DefaultKeyFunc
	}

	rateLimiter := NewRateLimiter(config.RequestsPerMinute, config.BurstSize)

	return func(c *gin.Context) {
		key := config.KeyFunc(c)
		
		if !rateLimiter.Allow(key) {
			c.Header("X-RateLimit-Limit", fmt.Sprintf("%d", config.RequestsPerMinute))
			c.Header("X-RateLimit-Remaining", "0")
			c.Header("X-RateLimit-Reset", fmt.Sprintf("%d", time.Now().Add(time.Minute).Unix()))
			
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":   "rate_limit_exceeded",
				"message": config.Message,
				"retry_after": "60",
			})
			c.Abort()
			return
		}

		c.Header("X-RateLimit-Limit", fmt.Sprintf("%d", config.RequestsPerMinute))
		c.Next()
	}
}

func RateLimitByIP(requestsPerMinute, burstSize int) gin.HandlerFunc {
	return RateLimit(&RateLimitConfig{
		RequestsPerMinute: requestsPerMinute,
		BurstSize:         burstSize,
		KeyFunc:           DefaultKeyFunc,
		Message:           "Too many requests from this IP",
	})
}

func RateLimitByUser(requestsPerMinute, burstSize int) gin.HandlerFunc {
	return RateLimit(&RateLimitConfig{
		RequestsPerMinute: requestsPerMinute,
		BurstSize:         burstSize,
		KeyFunc:           UserKeyFunc,
		Message:           "Too many requests from this user",
	})
}

func RateLimitAuth(requestsPerMinute, burstSize int) gin.HandlerFunc {
	return RateLimit(&RateLimitConfig{
		RequestsPerMinute: requestsPerMinute,
		BurstSize:         burstSize,
		KeyFunc:           DefaultKeyFunc,
		Message:           "Too many authentication attempts",
	})
}