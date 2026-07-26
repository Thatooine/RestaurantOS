package errs

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestHTTPStatus(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want int
	}{
		{name: "not found", err: ErrNotFound, want: http.StatusNotFound},
		{name: "wrapped forbidden", err: errors.Join(errors.New("operation failed"), ErrForbidden), want: http.StatusForbidden},
		{name: "conflict", err: ErrConflict, want: http.StatusConflict},
		{name: "unknown", err: errors.New("unexpected"), want: http.StatusInternalServerError},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := HTTPStatus(tt.err); got != tt.want {
				t.Fatalf("status: got %d, want %d", got, tt.want)
			}
		})
	}
}

func TestWriteGinError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	response := httptest.NewRecorder()
	context, _ := gin.CreateTestContext(response)

	WriteGinError(context, ErrNotFound)

	if response.Code != http.StatusNotFound {
		t.Fatalf("status: got %d, want %d", response.Code, http.StatusNotFound)
	}
	if got, want := response.Body.String(), `{"error":"not found"}`; got != want {
		t.Fatalf("body: got %q, want %q", got, want)
	}
}
