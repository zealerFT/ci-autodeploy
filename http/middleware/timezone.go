package middleware

import (
	"strconv"

	"github.com/gin-gonic/gin"
)

const (
	// KeyTimeOffset gin context key of time offset
	KeyTimeOffset = "timeOffset"
	// KeyTimezoneName gin context key of time offset
	KeyTimezoneName = "timezoneName"
)

var (
	defaultTimezoneOffsetHeaderKeys = []string{"x-timezone-offset", "X-Timezone-Offset"}
	defaultTimezoneNameHeaderKeys   = []string{"x-timezone-name", "X-Timezone-Name"}
)

// TimeOffset fetch time offset from header
func TimeOffset(headerKey ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if len(headerKey) == 0 {
			headerKey = defaultTimezoneOffsetHeaderKeys
		}

		val, ok := getHeader(c, headerKey...)

		if ok {
			offset, _ := strconv.Atoi(val)
			c.Set(KeyTimeOffset, offset)
		} else {
			c.Set(KeyTimeOffset, 0)
		}

		c.Next()
	}
}

// GetTimeOffset get time offset from context
func GetTimeOffset(c *gin.Context) int {
	return c.GetInt(KeyTimeOffset)
}

func TimezoneName(headerKey ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if len(headerKey) == 0 {
			headerKey = defaultTimezoneNameHeaderKeys
		}
		val, _ := getHeader(c, headerKey...)
		c.Set(KeyTimezoneName, val)

		c.Next()
	}
}

func GetTimezoneName(c *gin.Context) string {
	return c.GetString(KeyTimezoneName)
}
