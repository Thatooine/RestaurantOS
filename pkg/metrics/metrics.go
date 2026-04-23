package metrics

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

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

func init() {
	prometheus.MustRegister(requestsTotal, requestDuration, requestsInFlight)
}

// Handler returns an http.Handler that exposes Prometheus metrics.
func Handler() http.Handler {
	return promhttp.Handler()
}

// Middleware records request count, duration, and in-flight gauge for each request.
// It uses the gorilla/mux route template as the "route" label to keep cardinality bounded.
func Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestsInFlight.Inc()
		defer requestsInFlight.Dec()

		rw := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
		start := time.Now()

		next.ServeHTTP(rw, r)

		route := routeTemplate(r)
		requestDuration.WithLabelValues(r.Method, route).Observe(time.Since(start).Seconds())
		requestsTotal.WithLabelValues(r.Method, route, strconv.Itoa(rw.status)).Inc()
	})
}

func routeTemplate(r *http.Request) string {
	if current := mux.CurrentRoute(r); current != nil {
		if tmpl, err := current.GetPathTemplate(); err == nil {
			return tmpl
		}
	}
	return "unmatched"
}

type statusRecorder struct {
	http.ResponseWriter
	status      int
	wroteHeader bool
}

func (s *statusRecorder) WriteHeader(code int) {
	if s.wroteHeader {
		return
	}
	s.status = code
	s.wroteHeader = true
	s.ResponseWriter.WriteHeader(code)
}

func (s *statusRecorder) Write(b []byte) (int, error) {
	if !s.wroteHeader {
		s.wroteHeader = true
	}
	return s.ResponseWriter.Write(b)
}
