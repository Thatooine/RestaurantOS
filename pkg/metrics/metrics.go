package metrics

import (
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
)

// register collectors
func init() {
	prometheus.MustRegister(
		requestsTotal,
		requestDuration,
		requestsInFlight,
	)
}

var (
	requestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests processed, labeled by method, route, and status code.",
		},
		[]string{"method", "route", "status"},
	)

	requestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "Latency of HTTP requests in seconds, labeled by method and route.",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "route"},
	)

	requestsInFlight = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "http_requests_in_flight",
			Help: "Number of HTTP requests currently being served.",
		},
	)
)

// Middleware records request count, duration, and in-flight requests. Gin route
// patterns are normalized to the API's canonical label format so existing
// Prometheus series remain backward compatible and bounded.
func Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestsInFlight.Inc()
		defer requestsInFlight.Dec()

		start := time.Now()
		c.Next()

		route := canonicalRoute(c.FullPath())
		requestDuration.WithLabelValues(c.Request.Method, route).Observe(time.Since(start).Seconds())
		requestsTotal.WithLabelValues(c.Request.Method, route, strconv.Itoa(c.Writer.Status())).Inc()
	}
}

func canonicalRoute(route string) string {
	if route == "" {
		return "unmatched"
	}

	segments := strings.Split(route, "/")
	for i, segment := range segments {
		if len(segment) > 1 && (segment[0] == ':' || segment[0] == '*') {
			segments[i] = "{" + segment[1:] + "}"
		}
	}
	return strings.Join(segments, "/")
}
