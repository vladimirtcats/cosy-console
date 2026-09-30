package httptransport

import (
	"context"
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"cosy-console/internal/domain/orders"
	ordersmodels "cosy-console/internal/domain/orders/models"
	"cosy-console/pkg/pagination"
)

type orderUseCase interface {
	Get(ctx context.Context, id uuid.UUID) (ordersmodels.Order, error)
	ListByUser(
		ctx context.Context,
		userID int64,
		filter orders.OrderFilter,
		opts ...pagination.Option,
	) ([]ordersmodels.Order, pagination.Result, error)
	Create(ctx context.Context, params orders.CreationParams) (ordersmodels.Order, error)
	Cancel(ctx context.Context, id uuid.UUID) (ordersmodels.Order, error)
}

type OrderHandler struct {
	useCase orderUseCase
}

func NewOrderHandler(useCase orderUseCase) *OrderHandler {
	return &OrderHandler{
		useCase: useCase,
	}
}

type CreateOrderRequest struct {
	UserID int64             `json:"user_id" binding:"required"`
	Lines  []CreateOrderLine `json:"lines" binding:"required,min=1,dive"`
}

type CreateOrderLine struct {
	ItemID uuid.UUID `json:"item_id" binding:"required"`
	Qty    int       `json:"qty" binding:"required,min=1"`
}

func (h *OrderHandler) Get(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		respondBadRequest(c, "invalid order id")

		return
	}

	order, err := h.useCase.Get(c.Request.Context(), id)
	if err != nil {
		respondError(c, err)

		return
	}

	c.JSON(http.StatusOK, toOrderResponse(order))
}

func (h *OrderHandler) ListByUser(c *gin.Context) {
	userID, err := strconv.ParseInt(c.Param("user_id"), 10, 64)
	if err != nil {
		respondBadRequest(c, "invalid user id")

		return
	}

	opts, err := pageOpts(c)
	if err != nil {
		respondBadRequest(c, err.Error())

		return
	}

	ordersList, result, err := h.useCase.ListByUser(c.Request.Context(), userID, orders.OrderFilter{}, opts...)
	if err != nil {
		respondError(c, err)

		return
	}

	c.JSON(http.StatusOK, gin.H{
		"orders": toOrderResponses(ordersList),
		"page":   result,
	})
}

func (h *OrderHandler) Create(c *gin.Context) {
	var request CreateOrderRequest

	if err := c.ShouldBindJSON(&request); err != nil {
		respondBadRequest(c, err.Error())

		return
	}

	params := orders.CreationParams{
		UserID: request.UserID,
		Lines:  make([]orders.LineInput, 0, len(request.Lines)),
	}
	for _, line := range request.Lines {
		params.Lines = append(params.Lines, orders.LineInput{
			ItemID: line.ItemID,
			Qty:    line.Qty,
		})
	}

	order, err := h.useCase.Create(c.Request.Context(), params)
	if err != nil {
		respondError(c, err)

		return
	}

	c.JSON(http.StatusCreated, toOrderResponse(order))
}

func (h *OrderHandler) Cancel(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		respondBadRequest(c, "invalid order id")

		return
	}

	order, err := h.useCase.Cancel(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, orders.ErrInvalidTransition) {
			respond(c, http.StatusConflict, err.Error())

			return
		}

		respondError(c, err)

		return
	}

	c.JSON(http.StatusOK, toOrderResponse(order))
}
