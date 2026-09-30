package httptransport

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"cosy-console/internal/domain/catalog"
	"cosy-console/internal/domain/orders"
)

type errorResponse struct {
	Errors []string `json:"errors"`
}

func respondError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, orders.ErrNotFound), errors.Is(err, catalog.ErrNotFound):
		respond(c, http.StatusNotFound, "not found")
	default:
		respond(c, http.StatusInternalServerError, "internal server error")
	}
}

func respond(c *gin.Context, status int, messages ...string) {
	c.JSON(status, errorResponse{
		Errors: messages,
	})
}

func respondBadRequest(c *gin.Context, message string) {
	respond(c, http.StatusBadRequest, message)
}
