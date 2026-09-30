package httptransport

import (
	"errors"
	"strconv"

	"github.com/gin-gonic/gin"

	"cosy-console/pkg/pagination"
)

// pageOpts builds pagination options from the page and page_size query
// parameters; absent parameters keep the defaults.
func pageOpts(c *gin.Context) ([]pagination.Option, error) {
	opts := make([]pagination.Option, 0, 2)

	if raw := c.Query("page"); raw != "" {
		page, err := strconv.Atoi(raw)
		if err != nil || page < 1 {
			return nil, errors.New("page must be a positive integer")
		}

		opts = append(opts, pagination.WithPage(page))
	}

	if raw := c.Query("page_size"); raw != "" {
		pageSize, err := strconv.Atoi(raw)
		if err != nil || pageSize < 1 {
			return nil, errors.New("page_size must be a positive integer")
		}

		opts = append(opts, pagination.WithPageSize(pageSize))
	}

	return opts, nil
}
