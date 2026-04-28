package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

type RateLimiter struct {
	requests map[string][]*time.Time
	mu       sync.Mutex
	maxReq   int
	window   time.Duration
}

func NewRateLimiter(maxRequests int, window time.Duration) *RateLimiter {
	return &RateLimiter{
		requests: make(map[string][]*time.Time),
		maxReq:   maxRequests,
		window:   window,
	}
}

func (rl *RateLimiter) Limit() gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()

		rl.mu.Lock()
		defer rl.mu.Unlock()

		now := time.Now()
		if _, exists := rl.requests[ip]; !exists {
			rl.requests[ip] = make([]*time.Time, 0)
		}

		validRequests := make([]*time.Time, 0)
		for _, reqTime := range rl.requests[ip] {
			if now.Sub(*reqTime) < rl.window {
				validRequests = append(validRequests, reqTime)
			}
		}

		if len(validRequests) >= rl.maxReq {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error": "Trop de requêtes, réessayez plus tard.",
			})
			c.Abort()
			return
		}

		validRequests = append(validRequests, &now)
		rl.requests[ip] = validRequests
		c.Next()
	}
}
