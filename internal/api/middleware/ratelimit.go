package middleware

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

// RateLimiter provides rate limiting per agent
type RateLimiter struct {
	redis *redis.Client
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(redis *redis.Client) *RateLimiter {
	return &RateLimiter{redis: redis}
}

// Limit middleware limits requests per agent
func (r *RateLimiter) Limit() gin.HandlerFunc {
	return func(c *gin.Context) {
		agentID, exists := c.Get("agent_id")
		if !exists {
			c.Next()
			return
		}

		key := fmt.Sprintf("ratelimit:%v", agentID)

		// Sliding window counter using Redis
		now := time.Now().Unix()
		windowStart := now - 60 // 1 minute window

		pipe := r.redis.Pipeline()
		pipe.ZRemRangeByScore(c.Request.Context(), key, "0", fmt.Sprintf("%d", windowStart))
		pipe.ZCard(c.Request.Context(), key)
		pipe.ZAdd(c.Request.Context(), key, redis.Z{Score: float64(now), Member: now})
		pipe.Expire(c.Request.Context(), key, time.Minute)

		results, err := pipe.Exec(c.Request.Context())
		if err != nil {
			// Fail open — don't block on Redis error
			c.Next()
			return
		}

		count := results[1].(*redis.IntCmd).Val()

		// 1000 req/min — generous for any legitimate agent workflow
		if count > 1000 {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, ErrorResponse{
				Error:   "rate_limit",
				Message: "Slow down — you're making too many requests",
			})
			return
		}

		c.Next()
	}
}
