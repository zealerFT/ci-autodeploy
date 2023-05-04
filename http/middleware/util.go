package middleware

import (
	"context"

	"github.com/gin-gonic/gin"
)

func SetToContext(c *gin.Context, key string, value interface{}) {
	c.Set(key, value)
	ctx := context.WithValue(c.Request.Context(), key, value)
	c.Request = c.Request.WithContext(ctx)
}

func getHeader(c *gin.Context, keys ...string) (string, bool) {
	for _, key := range keys {
		if v := c.GetHeader(key); v != "" {
			return v, true
		}
	}
	return "", false
}
