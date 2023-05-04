package route

import (
	"autodeploy/http/api"
	"autodeploy/http/middleware"

	"github.com/gin-gonic/gin"
)

func AppAPI(opts ...func(engine *gin.Engine)) func(s *gin.Engine) {
	return func(s *gin.Engine) {
		for _, opt := range opts {
			opt(s)
		}
		// Not Found
		s.NoRoute(api.Handle404)
		// Health Check
		s.GET("/check", api.Health)
		// API 组
		SearchRoute := authRouteGroup(s, "/api/v1/autodeploy")
		// config get
		SearchRoute.POST("yaml/image/update", api.UpdateYamlImages)

	}
}

func authRouteGroup(s *gin.Engine, relativePath string) *gin.RouterGroup {
	group := s.Group(relativePath)
	group.Use(middleware.Auth())
	return group
}
