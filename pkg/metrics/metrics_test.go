package metrics

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
	dto "github.com/prometheus/client_model/go"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func TestCanonicalRoute(t *testing.T) {
	tests := []struct {
		path string
		want string
	}{
		{path: "", want: "unmatched"},
		{path: "/health", want: "/health"},
		{path: "/api/v1/dishes/:id", want: "/api/v1/dishes/{id}"},
		{path: "/api/v1/users/:email", want: "/api/v1/users/{email}"},
		{path: "/files/*path", want: "/files/{path}"},
	}

	for _, tt := range tests {
		if got := canonicalRoute(tt.path); got != tt.want {
			t.Errorf("canonicalRoute(%q): got %q, want %q", tt.path, got, tt.want)
		}
	}
}

func TestMiddlewareUsesCanonicalBoundedRouteLabels(t *testing.T) {
	engine := gin.New()
	engine.Use(Middleware())
	engine.GET("/items/:id", func(c *gin.Context) {
		c.Status(http.StatusNoContent)
	})

	counter := requestsTotal.WithLabelValues(http.MethodGet, "/items/{id}", "204")
	before := testutil.ToFloat64(counter)
	inFlightBefore := testutil.ToFloat64(requestsInFlight)

	request := httptest.NewRequest(http.MethodGet, "/items/concrete-id", nil)
	response := httptest.NewRecorder()
	engine.ServeHTTP(response, request)

	if response.Code != http.StatusNoContent {
		t.Fatalf("status: got %d, want %d", response.Code, http.StatusNoContent)
	}
	if got := testutil.ToFloat64(counter); got != before+1 {
		t.Fatalf("canonical counter: got %f, want %f", got, before+1)
	}
	if got := testutil.ToFloat64(requestsInFlight); got != inFlightBefore {
		t.Fatalf("in-flight gauge after request: got %f, want %f", got, inFlightBefore)
	}
	assertNoRouteLabel(t, "/items/concrete-id")
	assertNoRouteLabel(t, "/items/:id")
}

func TestMiddlewareRecordsUnmatchedAndRecoveredRequests(t *testing.T) {
	t.Run("unmatched", func(t *testing.T) {
		engine := gin.New()
		engine.Use(Middleware())

		counter := requestsTotal.WithLabelValues(http.MethodGet, "unmatched", "404")
		before := testutil.ToFloat64(counter)

		response := httptest.NewRecorder()
		engine.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/does-not-exist", nil))

		if response.Code != http.StatusNotFound {
			t.Fatalf("status: got %d, want %d", response.Code, http.StatusNotFound)
		}
		if got := testutil.ToFloat64(counter); got != before+1 {
			t.Fatalf("unmatched counter: got %f, want %f", got, before+1)
		}
		assertNoRouteLabel(t, "/does-not-exist")
	})

	t.Run("recovered panic", func(t *testing.T) {
		engine := gin.New()
		engine.Use(Middleware(), gin.RecoveryWithWriter(io.Discard))
		engine.GET("/panic/:id", func(_ *gin.Context) {
			panic("boom")
		})

		counter := requestsTotal.WithLabelValues(http.MethodGet, "/panic/{id}", "500")
		before := testutil.ToFloat64(counter)
		inFlightBefore := testutil.ToFloat64(requestsInFlight)

		response := httptest.NewRecorder()
		engine.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/panic/concrete-id", nil))

		if response.Code != http.StatusInternalServerError {
			t.Fatalf("status: got %d, want %d", response.Code, http.StatusInternalServerError)
		}
		if got := testutil.ToFloat64(counter); got != before+1 {
			t.Fatalf("recovered counter: got %f, want %f", got, before+1)
		}
		if got := testutil.ToFloat64(requestsInFlight); got != inFlightBefore {
			t.Fatalf("in-flight gauge after panic: got %f, want %f", got, inFlightBefore)
		}
		assertNoRouteLabel(t, "/panic/concrete-id")
	})
}

func assertNoRouteLabel(t *testing.T, unwanted string) {
	t.Helper()
	families, err := prometheus.DefaultGatherer.Gather()
	if err != nil {
		t.Fatalf("gather metrics: %v", err)
	}
	for _, family := range families {
		if family.GetName() != "http_requests_total" && family.GetName() != "http_request_duration_seconds" {
			continue
		}
		for _, metric := range family.Metric {
			if labelValue(metric, "route") == unwanted {
				t.Fatalf("metric %s contains unbounded route label %q", family.GetName(), unwanted)
			}
		}
	}
}

func labelValue(metric *dto.Metric, name string) string {
	for _, label := range metric.Label {
		if label.GetName() == name {
			return label.GetValue()
		}
	}
	return ""
}
