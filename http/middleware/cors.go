package middleware

import (
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

var defaultCORSConfig = cors.Config{
	AllowAllOrigins: true,
	AllowMethods:    []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS", "HEAD"},
	AllowHeaders: []string{
		"Origin",
		"Accept",
		"Accept-Language",
		"Content-Language",
		"Content-Type",
		"User-Agent",
		"Authorization",
		"x-timezone-offset",
		"x-user-id",
		"X-User-Id",
		"X-Rate-Limit-Token",
		"x-rate-limit-token",
		"X-Timezone-Name",
		"x-timezone-name",
		"x-user-language",
	},
	AllowCredentials: true,
	MaxAge:           12 * time.Hour,
}

// CORS enable CORS support
func CORS(configs ...cors.Config) gin.HandlerFunc {
	if len(configs) != 0 {
		return cors.New(configs[0])
	}

	return cors.New(defaultCORSConfig)
}
