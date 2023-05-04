package prometheussvc

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"
)

var subsystem = "http"

var metricReqTotal = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Subsystem: subsystem,
		Name:      "requests_total",
		Help:      "How many HTTP requests processed, partitioned by status code and HTTP method.",
	},
	[]string{"code", "method", "path"},
)

var metricReqDuration = prometheus.NewHistogramVec(
	prometheus.HistogramOpts{
		Subsystem: subsystem,
		Name:      "request_duration_seconds",
		Help:      "The HTTP request latencies in seconds.",
	},
	[]string{"method", "path"},
)

var metricReqSize = prometheus.NewHistogramVec(
	prometheus.HistogramOpts{
		Subsystem: subsystem,
		Name:      "request_size_bytes",
		Help:      "The HTTP request sizes in bytes.",
	},
	[]string{"method", "path"},
)

var metricResSize = prometheus.NewHistogramVec(
	prometheus.HistogramOpts{
		Subsystem: subsystem,
		Name:      "response_size_bytes",
		Help:      "The HTTP response sizes in bytes.",
	},
	[]string{"method", "path"},
)

// Prometheus contains the metrics gathered by the instance and its path
type Prometheus struct {
	router        *gin.Engine
	listenAddress string
	Gateway       PrometheusPushGateway

	MetricsPath string
	HealthPath  string
}

// PrometheusPushGateway contains the configuration for pushing to a Prometheus pushgateway (optional)
type PrometheusPushGateway struct {
	// Push interval in seconds
	PushIntervalSeconds time.Duration

	// Push Gateway URL in format http://domain:port
	// where JOBNAME can be any string of your choice
	PushGatewayURL string

	// Local metrics URL where metrics are fetched from, this could be ommited in the future
	// if implemented using prometheus common/expfmt instead
	MetricsURL string

	// pushgateway job name, defaults to "gin"
	Job string
}

// NewPrometheus generates a new set of metrics with a certain subsystem name
func NewPrometheus() *Prometheus {

	p := &Prometheus{
		MetricsPath: "/metrics",
		HealthPath:  "/health",
	}
	p.RegisterMetrics(metricReqTotal, metricReqDuration, metricReqSize, metricResSize)

	return p
}

// SetPushGateway sends metrics to a remote pushgateway exposed on pushGatewayURL
// every pushIntervalSeconds. Metrics are fetched from metricsURL
func (p *Prometheus) SetPushGateway(pushGatewayURL, metricsURL string, pushIntervalSeconds time.Duration) {
	p.Gateway.PushGatewayURL = pushGatewayURL
	p.Gateway.MetricsURL = metricsURL
	p.Gateway.PushIntervalSeconds = pushIntervalSeconds
	p.startPushTicker()
}

// SetPushGatewayJob job name, defaults to "gin"
func (p *Prometheus) SetPushGatewayJob(j string) {
	p.Gateway.Job = j
}

// SetListenAddress for exposing metrics on address. If not set, it will be exposed at the
// same address of the gin engine that is being used
func (p *Prometheus) SetListenAddress(address string) {
	p.listenAddress = address
	if p.listenAddress != "" {
		p.router = gin.Default()
	}
}

// SetListenAddressWithRouter for using a separate router to expose metrics. (this keeps things like GET /metrics out of
// your content's access log).
func (p *Prometheus) SetListenAddressWithRouter(listenAddress string, r *gin.Engine) {
	p.listenAddress = listenAddress
	if len(p.listenAddress) > 0 {
		p.router = r
	}
}

// SetMetricsPath set metrics paths
func (p *Prometheus) SetMetricsPath(e *gin.Engine) {

	if p.listenAddress != "" {
		p.router.GET(p.MetricsPath, prometheusHandler())
		p.runServer()
	} else {
		e.GET(p.MetricsPath, prometheusHandler())
	}
}

// SetMetricsPathWithAuth set metrics paths with authentication
func (p *Prometheus) SetMetricsPathWithAuth(e *gin.Engine, accounts gin.Accounts) {

	if p.listenAddress != "" {
		p.router.GET(p.MetricsPath, gin.BasicAuth(accounts), prometheusHandler())
		p.runServer()
	} else {
		e.GET(p.MetricsPath, gin.BasicAuth(accounts), prometheusHandler())
	}

}

func (p *Prometheus) runServer() {
	if p.listenAddress != "" {
		go p.router.Run(p.listenAddress)
	}
}

func (p *Prometheus) getMetrics() []byte {
	response, _ := http.Get(p.Gateway.MetricsURL)

	defer response.Body.Close()
	body, _ := ioutil.ReadAll(response.Body)

	return body
}

func (p *Prometheus) getPushGatewayURL() string {
	h, _ := os.Hostname()
	if p.Gateway.Job == "" {
		p.Gateway.Job = "gin"
	}
	return p.Gateway.PushGatewayURL + "/metrics/job/" + p.Gateway.Job + "/instance/" + h
}

func (p *Prometheus) sendMetricsToPushGateway(metrics []byte) {
	req, err := http.NewRequest("POST", p.getPushGatewayURL(), bytes.NewBuffer(metrics))
	client := &http.Client{}
	if _, err = client.Do(req); err != nil {
		log.WithError(err).Errorln("Error sending to push gateway")
	}
}

func (p *Prometheus) startPushTicker() {
	ticker := time.NewTicker(time.Second * p.Gateway.PushIntervalSeconds)
	go func() {
		for range ticker.C {
			p.sendMetricsToPushGateway(p.getMetrics())
		}
	}()
}

// RegisterMetrics Customizable metrics registration is required before server startup
func (p *Prometheus) RegisterMetrics(controllers ...prometheus.Collector) {
	for _, m := range controllers {
		if err := prometheus.Register(m); err != nil {
			log.WithError(err).Errorf("could not be registered in Prometheus, check your Metrics")
		}
	}
}

// Use adds the middleware to a gin engine.
func (p *Prometheus) Use(e *gin.Engine) {
	e.Use(p.HandlerFunc(e))
	e.Use(gin.LoggerWithConfig(gin.LoggerConfig{SkipPaths: []string{"/metrics", "/health"}}))
	p.SetMetricsPath(e)
}

// UseWithAuth adds the middleware to a gin engine with BasicAuth.
func (p *Prometheus) UseWithAuth(e *gin.Engine, accounts gin.Accounts) {
	e.Use(p.HandlerFunc(e))
	e.Use(gin.LoggerWithConfig(gin.LoggerConfig{SkipPaths: []string{"/metrics", "/health"}}))
	p.SetMetricsPathWithAuth(e, accounts)
}

// HandlerFunc defines handler function for middleware
func (p *Prometheus) HandlerFunc(e *gin.Engine) gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.FullPath() == "" || c.Request.URL.Path == p.MetricsPath || c.Request.URL.Path == p.HealthPath {
			c.Next()
			return
		}

		start := time.Now()
		reqSize := float64(computeApproximateRequestSize(c.Request))

		c.Next()

		status := strconv.Itoa(c.Writer.Status())
		elapsed := float64(time.Since(start)) / float64(time.Second)
		resSize := float64(c.Writer.Size())

		fullPath := c.FullPath()

		metricReqTotal.WithLabelValues(status, c.Request.Method, fullPath).Inc()
		metricReqDuration.WithLabelValues(c.Request.Method, fullPath).Observe(elapsed)
		metricReqSize.WithLabelValues(c.Request.Method, fullPath).Observe(reqSize)
		metricResSize.WithLabelValues(c.Request.Method, fullPath).Observe(resSize)
	}
}

func prometheusHandler() gin.HandlerFunc {
	h := promhttp.Handler()
	return func(c *gin.Context) {
		h.ServeHTTP(c.Writer, c.Request)
	}
}

// From https://github.com/DanielHeckrath/gin-prometheus/blob/master/gin_prometheus.go
func computeApproximateRequestSize(r *http.Request) int {
	s := 0
	if r.URL != nil {
		s = len(r.URL.Path)
	}

	s += len(r.Method)
	s += len(r.Proto)
	for name, values := range r.Header {
		s += len(name)
		for _, value := range values {
			s += len(value)
		}
	}
	s += len(r.Host)

	// N.B. r.Form and r.MultipartForm are assumed to be included in r.URL.

	if r.ContentLength != -1 {
		s += int(r.ContentLength)
	}
	return s
}
