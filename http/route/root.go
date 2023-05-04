package route

import (
	"autodeploy/http/middleware"
	m "autodeploy/http/middleware"

	"github.com/gin-contrib/gzip"
	"github.com/gin-contrib/requestid"
	"github.com/gin-gonic/gin"
)

func Root(opts ...func(engine *gin.Engine)) func(s *gin.Engine) {
	return func(s *gin.Engine) {
		for _, opt := range opts {
			opt(s)
		}

		// common middleware
		s.Use(
			m.CORS(),
			m.TimeOffset("x-timezone-offset"),
			m.TimezoneName(),
			middleware.Pagination(),
			requestid.New(),
			gzip.Gzip(gzip.DefaultCompression),
		)
	}
}
