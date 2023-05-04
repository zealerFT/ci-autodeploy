package prometheussvc

import (
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
)

var (
	totalReqsCounter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Subsystem: "jarvis",
			Name:      "total_deploy_reqs",
			Help:      "Total requests by team, app, author",
		},
		[]string{"team", "app", "author"},
	)
)

type MetricsList struct {
	TotalReqsCounter *prometheus.CounterVec
}

func MustRegister(engine *gin.Engine, addr string) *MetricsList {
	p := NewPrometheus()
	p.RegisterMetrics(totalReqsCounter)
	p.SetListenAddress(addr)
	p.Use(engine)
	return &MetricsList{
		TotalReqsCounter: totalReqsCounter,
	}
}
