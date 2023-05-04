package http

import (
	"net/http"
	"strings"

	"autodeploy/http/route"
	. "autodeploy/util/object"
	"autodeploy/util/prometheussvc"

	"github.com/gin-gonic/gin"
)

type Server struct {
	*gin.Engine
}

func New(options ...Option) *Server {
	s := &Server{
		Engine: gin.New(),
	}

	for _, option := range options {
		option(s)
	}

	return s
}

type Option func(*Server)

// ExportLogOption 需要跳过日志记录的情况 "/metrics", "/check" 这两个接口是用来pod检查和metrics采集的
func ExportLogOption() Option {
	return func(s *Server) {
		s.Engine.Use(gin.LoggerWithConfig(gin.LoggerConfig{SkipPaths: []string{"/metrics", "/check"}}))
	}
}

func SetRouteOption() Option {
	return func(s *Server) {
		route.HTTPServerRoute()(s.Engine)
	}
}

func SetPrometheusMetrics() Option {
	return func(s *Server) {
		s.Engine.Use(P(Metrics.Set, prometheussvc.MustRegister(s.Engine, ":23333")))
	}
}

func WhitelistCheck() Option {
	return func(s *Server) {
		s.Engine.Use( // 定义白名单
			// 定义中间件
			func(c *gin.Context) {
				// 判断是否是/metrics请求，如果是，则直接跳过
				if strings.HasPrefix(c.Request.URL.Path, "/metrics") || strings.HasPrefix(c.Request.URL.Path, "/check") || strings.HasPrefix(c.Request.URL.Path, "/health") {
					c.Next()
					return
				}
				whitelist := []string{
					"127.0.0.1",
					"::1", // 这里是本地测试的白名单，正式上线需要加入ci-runner的公网ip，这样加更加安全
				}
				// 获取请求的IP地址
				ip := c.ClientIP()
				// 检查是否在白名单中
				for _, allowedIP := range whitelist {
					if allowedIP == ip || strings.HasPrefix(allowedIP, ip+":") {
						c.Next()
						return
					}
				}
				// 如果不在白名单中则返回错误
				c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized ip"})
				c.Abort()
			})
	}
}
