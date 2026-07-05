package middleware

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/user/network-monitoring/internal/repository"
)

func RateLimiter(limit int) gin.HandlerFunc {
	return func(c *gin.Context) {
		if repository.RDB == nil {
			// Redis not initialized, bypass rate limit
			c.Next()
			return
		}

		ctx := context.Background()
		clientIP := c.ClientIP()
		key := fmt.Sprintf("rate:%s:%d", clientIP, time.Now().Minute())

		count, err := repository.RDB.Incr(ctx, key).Result()
		if err != nil {
			// Fail open on cache read failures
			c.Next()
			return
		}

		if count == 1 {
			repository.RDB.Expire(ctx, key, 90*time.Second)
		}

		if count > int64(limit) {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error": "Too many requests. Please try again later.",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}
