package handlers

import (
	"net/http"
	"order-service/internal/models"
	"order-service/internal/service"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	service service.OrderService
}

func New(service service.OrderService) *Handler {
	return &Handler{
		service: service,
	}
}

func (h *Handler) GetOrder(c *gin.Context) {
	orderID := c.Param("id")
	if orderID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Order ID is required"})
		return
	}

	order, err := h.service.GetOrder(orderID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Order not found"})
		return
	}

	c.JSON(http.StatusOK, order)
}

func (h *Handler) WebInterface(c *gin.Context) {
	c.HTML(http.StatusOK, "index.html", nil)
}

func (h *Handler) HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":    "healthy",
		"cacheSize": h.service.GetCacheSize(),
	})
}