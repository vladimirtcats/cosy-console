package app

import (
	"net/http"

	"github.com/gin-gonic/gin"

	httptransport "cosy-console/internal/transport/http"
)

func NewRouter(container *Container) *gin.Engine {
	router := gin.New()
	router.Use(gin.Recovery())

	orderHandler := httptransport.NewOrderHandler(container.Orders.Usecases.Order)
	itemHandler := httptransport.NewItemHandler(container.Catalog.Usecases.Item)

	api := router.Group("/api/v1")
	{
		api.GET("/orders/:id", orderHandler.Get)
		api.POST("/orders", orderHandler.Create)
		api.POST("/orders/:id/cancel", orderHandler.Cancel)
		api.GET("/users/:user_id/orders", orderHandler.ListByUser)
		api.GET("/items", itemHandler.List)
	}

	router.GET("/health", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	return router
}
