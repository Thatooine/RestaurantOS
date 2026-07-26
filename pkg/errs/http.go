package errs

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
)

// HTTPStatus maps application sentinel errors to their public HTTP status.
func HTTPStatus(err error) int {
	switch {
	case errors.Is(err, ErrNotFound):
		return http.StatusNotFound
	case errors.Is(err, ErrForbidden):
		return http.StatusForbidden
	case errors.Is(err, ErrConflict):
		return http.StatusConflict
	default:
		return http.StatusInternalServerError
	}
}

// WriteGinError writes the API's standard JSON error response.
func WriteGinError(c *gin.Context, err error) {
	c.JSON(HTTPStatus(err), gin.H{"error": err.Error()})
}
