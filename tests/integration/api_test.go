package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"strings"
	"testing"

	"github.com/redis/go-redis/v9"
)

const defaultBaseURL = "http://localhost:8080"

func baseURL() string {
	if v := os.Getenv("API_BASE_URL"); v != "" {
		return v
	}
	return defaultBaseURL
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type loginResponse struct {
	Token  string `json:"token"`
	UserID string `json:"userID"`
	Email  string `json:"email"`
}

type dish struct {
	ID           string  `json:"id"`
	RestaurantID string  `json:"restaurant_id"`
	Name         string  `json:"name"`
	Description  string  `json:"description"`
	Price        float64 `json:"price"`
	Image        string  `json:"image"`
}

type listDishesResponse struct {
	Dishes []dish `json:"dishes"`
	Total  int64  `json:"total"`
}

func loginAsRootUser(t *testing.T) string {
	t.Helper()

	body, err := json.Marshal(loginRequest{
		Email:    "root+user@gmail.com",
		Password: "abc123",
	})
	if err != nil {
		t.Fatalf("failed to marshal login request: %v", err)
	}

	resp, err := http.Post(baseURL()+"/api/v1/auth/login", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("failed to send login request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected status 200 on login, got %d", resp.StatusCode)
	}

	var loginResp loginResponse
	if err := json.NewDecoder(resp.Body).Decode(&loginResp); err != nil {
		t.Fatalf("failed to decode login response: %v", err)
	}

	if loginResp.Token == "" {
		t.Fatal("login response returned empty token")
	}

	return loginResp.Token
}

func TestLoginAndListDishes(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	token := loginAsRootUser(t)

	req, err := http.NewRequest(http.MethodGet, baseURL()+"/api/v1/dishes", nil)
	if err != nil {
		t.Fatalf("failed to create list dishes request: %v", err)
	}
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("failed to send list dishes request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected status 200 on list dishes, got %d", resp.StatusCode)
	}

	var dishesResp listDishesResponse
	if err := json.NewDecoder(resp.Body).Decode(&dishesResp); err != nil {
		t.Fatalf("failed to decode list dishes response: %v", err)
	}

	if dishesResp.Total == 0 {
		t.Fatal("expected at least one dish, got 0")
	}

	if len(dishesResp.Dishes) == 0 {
		t.Fatal("dishes array is empty")
	}

	for _, d := range dishesResp.Dishes {
		if d.ID == "" {
			t.Error("dish has empty ID")
		}
		if d.Name == "" {
			t.Error("dish has empty name")
		}
	}

	const seededDishID = "00000000-0000-0000-0000-000000000101"
	getRequest, err := http.NewRequest(http.MethodGet, baseURL()+"/api/v1/dishes/"+seededDishID, nil)
	if err != nil {
		t.Fatalf("failed to create get dish request: %v", err)
	}
	getRequest.Header.Set("Authorization", "Bearer "+token)

	getResponse, err := http.DefaultClient.Do(getRequest)
	if err != nil {
		t.Fatalf("failed to send get dish request: %v", err)
	}
	defer getResponse.Body.Close()
	if getResponse.StatusCode != http.StatusOK {
		t.Fatalf("expected status 200 on get dish, got %d", getResponse.StatusCode)
	}

	metricsResponse, err := http.Get(baseURL() + "/metrics")
	if err != nil {
		t.Fatalf("failed to fetch metrics: %v", err)
	}
	defer metricsResponse.Body.Close()
	metricsBody, err := io.ReadAll(metricsResponse.Body)
	if err != nil {
		t.Fatalf("failed to read metrics: %v", err)
	}
	metrics := string(metricsBody)
	if !strings.Contains(metrics, `http_requests_total{method="GET",route="/api/v1/dishes/{id}",status="200"}`) {
		t.Fatal("metrics do not contain the canonical parameterized dish route label")
	}
	if strings.Contains(metrics, `route="/api/v1/dishes/`+seededDishID+`"`) || strings.Contains(metrics, `route="/api/v1/dishes/:id"`) {
		t.Fatal("metrics contain a concrete or Gin-native dish route label")
	}

	t.Logf("listed %d dishes (total: %d)", len(dishesResp.Dishes), dishesResp.Total)
}

func TestRouterHTTPContract(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	tests := []struct {
		name       string
		method     string
		path       string
		wantStatus int
		wantBody   string
	}{
		{name: "health", method: http.MethodGet, path: "/health", wantStatus: http.StatusOK, wantBody: `{"status":"ok"}`},
		{name: "trailing slash is not redirected", method: http.MethodGet, path: "/health/", wantStatus: http.StatusNotFound, wantBody: "404 page not found\n"},
		{name: "wrong method", method: http.MethodPost, path: "/health", wantStatus: http.StatusMethodNotAllowed},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request, err := http.NewRequest(tt.method, baseURL()+tt.path, nil)
			if err != nil {
				t.Fatalf("failed to create request: %v", err)
			}
			response, err := http.DefaultClient.Do(request)
			if err != nil {
				t.Fatalf("failed to send request: %v", err)
			}
			defer response.Body.Close()
			body, err := io.ReadAll(response.Body)
			if err != nil {
				t.Fatalf("failed to read response: %v", err)
			}
			if response.StatusCode != tt.wantStatus {
				t.Fatalf("status: got %d, want %d", response.StatusCode, tt.wantStatus)
			}
			if string(body) != tt.wantBody {
				t.Fatalf("body: got %q, want %q", string(body), tt.wantBody)
			}
			if allow := response.Header.Get("Allow"); allow != "" {
				t.Fatalf("unexpected Allow header: %q", allow)
			}
		})
	}
}

func redisAddr() string {
	if v := os.Getenv("REDIS_ADDR"); v != "" {
		return v
	}
	return "localhost:6379"
}

// flushIPRateLimitKeys deletes all IP rate limit keys from Redis so the test
// starts with a full token bucket.
func flushIPRateLimitKeys(t *testing.T) {
	t.Helper()

	rdb := redis.NewClient(&redis.Options{Addr: redisAddr()})
	defer rdb.Close()

	ctx := context.Background()
	// mennanov/limiters wraps the configured key in a Redis hash-tag and adds a
	// state suffix, for example: {ip_token_bucket:::1}state.
	keys, err := rdb.Keys(ctx, "{ip_token_bucket:*").Result()
	if err != nil {
		t.Fatalf("failed to scan Redis for IP rate limit keys: %v", err)
	}
	if len(keys) > 0 {
		if err := rdb.Del(ctx, keys...).Err(); err != nil {
			t.Fatalf("failed to delete IP rate limit keys: %v", err)
		}
	}
}

func TestIPRateLimitOnLogin(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	// Clean slate: remove any existing IP rate limit state.
	flushIPRateLimitKeys(t)

	body, err := json.Marshal(
		loginRequest{
			Email:    "root+user@gmail.com",
			Password: "abc123",
		},
	)
	if err != nil {
		t.Fatalf("failed to marshal login request: %v", err)
	}

	loginURL := baseURL() + "/api/v1/auth/login"

	// The server is configured with capacity=5 and refill=1 token/minute.
	// Send 5 requests that should all succeed, then the 6th should be rate-limited.
	const bucketCapacity = 5

	for i := 1; i <= bucketCapacity; i++ {
		resp, err := http.Post(loginURL, "application/json", bytes.NewReader(body))
		if err != nil {
			t.Fatalf("request %d: failed to send: %v", i, err)
		}
		resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Fatalf("request %d: expected 200, got %d", i, resp.StatusCode)
		}
	}

	// The 6th request should exceed the rate limit.
	resp, err := http.Post(loginURL, "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("rate-limited request: failed to send: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusTooManyRequests {
		t.Fatalf("expected status 429 after %d requests, got %d", bucketCapacity+1, resp.StatusCode)
	}

	var errResp map[string]string
	if err := json.NewDecoder(resp.Body).Decode(&errResp); err != nil {
		t.Fatalf("failed to decode error response: %v", err)
	}

	if errResp["error"] != "too many requests" {
		t.Errorf("expected error \"too many requests\", got %q", errResp["error"])
	}

	t.Logf("IP rate limit triggered after %d requests as expected", bucketCapacity+1)
}

// flushUserRateLimitKeys deletes all user-based rate limit keys from Redis so
// the test starts with a full token bucket.
func flushUserRateLimitKeys(t *testing.T) {
	t.Helper()

	rdb := redis.NewClient(&redis.Options{Addr: redisAddr()})
	defer rdb.Close()

	ctx := context.Background()
	// mennanov/limiters wraps the configured key in a Redis hash-tag and adds a
	// state suffix, for example: {token_bucket:user-id}state.
	keys, err := rdb.Keys(ctx, "{token_bucket:*").Result()
	if err != nil {
		t.Fatalf("failed to scan Redis for user rate limit keys: %v", err)
	}
	if len(keys) > 0 {
		if err := rdb.Del(ctx, keys...).Err(); err != nil {
			t.Fatalf("failed to delete user rate limit keys: %v", err)
		}
	}
}

func TestUserRateLimitOnAuthenticatedRoutes(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	// Clean slate: remove any existing user and IP rate limit state.
	flushUserRateLimitKeys(t)
	flushIPRateLimitKeys(t)

	token := loginAsRootUser(t)

	dishesURL := baseURL() + "/api/v1/dishes"

	// The server is configured with capacity=20 and refill=1 token/second.
	// Send 20 requests that should all succeed, then the 21st should be rate-limited.
	const bucketCapacity = 20

	for i := 1; i <= bucketCapacity; i++ {
		req, err := http.NewRequest(http.MethodGet, dishesURL, nil)
		if err != nil {
			t.Fatalf("request %d: failed to create request: %v", i, err)
		}
		req.Header.Set("Authorization", "Bearer "+token)

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			t.Fatalf("request %d: failed to send: %v", i, err)
		}
		resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			t.Fatalf("request %d: expected 200, got %d", i, resp.StatusCode)
		}
	}

	// The 21st request should exceed the user rate limit.
	req, err := http.NewRequest(http.MethodGet, dishesURL, nil)
	if err != nil {
		t.Fatalf("rate-limited request: failed to create request: %v", err)
	}
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("rate-limited request: failed to send: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusTooManyRequests {
		t.Fatalf("expected status 429 after %d requests, got %d", bucketCapacity+1, resp.StatusCode)
	}

	var errResp map[string]string
	if err := json.NewDecoder(resp.Body).Decode(&errResp); err != nil {
		t.Fatalf("failed to decode error response: %v", err)
	}

	if errResp["error"] != "too many requests" {
		t.Errorf("expected error \"too many requests\", got %q", errResp["error"])
	}

	t.Logf("user rate limit triggered after %d requests as expected", bucketCapacity+1)
}
