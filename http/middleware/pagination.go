package middleware

import (
	"fmt"
	"strconv"

	"github.com/gin-gonic/gin"
)

// Pagination sets page, pageSize and pageOffset to *gin.Context
func Pagination() gin.HandlerFunc {
	return func(c *gin.Context) {
		page := getSet(c, "page", 1)
		size := getSet(c, "pageSize", 20)
		c.Set("pageOffset", (page-1)*size)
		c.Next()
	}
}

func getSet(c *gin.Context, k string, d int) int {
	var n int
	if v := c.Query(k); v != "" {
		if i, err := strconv.Atoi(v); err == nil {
			if i > 0 {
				n = i
			}
		}
	}

	if n == 0 {
		n = d
	}

	c.Set(k, n)
	c.Request.URL.Query().Set(k, fmt.Sprintf("%d", n))
	return n
}
