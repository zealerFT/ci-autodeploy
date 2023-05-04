package route

import (
	"github.com/gin-gonic/gin"
)

// HTTPServerRoute 核心 App 服务 HTTP 路由
func HTTPServerRoute(opts ...func(engine *gin.Engine)) func(s *gin.Engine) {
	routes := []func(s *gin.Engine){
		Root(opts...),
		AppAPI(opts...),
	}

	return func(s *gin.Engine) {
		for _, route := range routes {
			route(s)
		}
	}
}
