package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"
)

func Auth() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := authToken(c)
		if token != "12345678" { // 这里是简单的token鉴权，根据需要来修改成复杂的token
			c.AbortWithStatusJSON(401, "no auth!")
			return
		}
		c.Next()
	}
}

// authToken 获取 鉴权 Bearer authToken
func authToken(c *gin.Context) string {
	value := ginHeader(c, "Authorization")
	if value != "" {
		return trimPrefix(value, "Bearer ")
	}

	return ""
}

func ginHeader(c *gin.Context, k string) string {
	if value := c.GetHeader(k); value != "" {
		return value
	}

	return c.GetHeader(strings.ToLower(k))
}

func trimPrefix(value, prefix string) string {
	if value == "" {
		return value
	}

	if strings.HasPrefix(value, prefix) {
		return strings.TrimPrefix(value, prefix)
	}

	if strings.HasPrefix(value, strings.ToLower(prefix)) {
		return strings.TrimPrefix(value, strings.ToLower(prefix))
	}

	return ""
}
