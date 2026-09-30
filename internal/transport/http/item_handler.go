package httptransport

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"

	catalogdomain "cosy-console/internal/domain/catalog"
	catalogmodels "cosy-console/internal/domain/catalog/models"
	"cosy-console/pkg/pagination"
)

type itemUseCase interface {
	List(
		ctx context.Context,
		filter catalogdomain.ItemFilter,
		opts ...pagination.Option,
	) ([]catalogmodels.Item, pagination.Result, error)
}

type ItemHandler struct {
	useCase itemUseCase
}

func NewItemHandler(useCase itemUseCase) *ItemHandler {
	return &ItemHandler{
		useCase: useCase,
	}
}

func (h *ItemHandler) List(c *gin.Context) {
	opts, err := pageOpts(c)
	if err != nil {
		respondBadRequest(c, err.Error())

		return
	}

	filter := catalogdomain.ItemFilter{
		SKU:             c.Query("sku"),
		IncludeInactive: c.Query("include_inactive") == "true",
	}

	items, result, err := h.useCase.List(c.Request.Context(), filter, opts...)
	if err != nil {
		respondError(c, err)

		return
	}

	c.JSON(http.StatusOK, gin.H{
		"items": toItemResponses(items),
		"page":  result,
	})
}
