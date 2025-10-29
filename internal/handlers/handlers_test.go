package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"order-service/internal/models"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

type mockService struct {
	order *models.Order
	err   error
}

func (m *mockService) ProcessMessage(data []byte) error {
	return m.err
}

func (m *mockService) GetOrder(orderUID string) (*models.Order, error) {
	return m.order, m.err
}

func (m *mockService) GetCacheSize() int {
	if m.order != nil {
		return 1
	}
	return 0
}

func (m *mockService) RestoreCache() error {
	return m.err
}

func TestGetOrderHandler(t *testing.T) {
	testOrder := &models.Order{
		OrderUID:    "test-order-123",
		TrackNumber: "WBILMTESTTRACK",
		Entry:       "WBIL",
		Locale:      "en",
		CustomerID:  "test",
		DateCreated: time.Now(),
	}

	mockSvc := &mockService{order: testOrder}
	handler := New(mockSvc)

	router := gin.New()
	router.GET("/api/order/:id", handler.GetOrder)

	req, err := http.NewRequest("GET", "/api/order/test-order-123", nil)
	assert.NoError(t, err)

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var response models.Order
	err = json.Unmarshal(rr.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, testOrder.OrderUID, response.OrderUID)
}

func TestGetOrderHandlerNotFound(t *testing.T) {
	mockSvc := &mockService{order: nil, err: assert.AnError}
	handler := New(mockSvc)

	router := gin.New()
	router.GET("/api/order/:id", handler.GetOrder)

	req, err := http.NewRequest("GET", "/api/order/nonexistent", nil)
	assert.NoError(t, err)

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusNotFound, rr.Code)
}

func TestGetOrderHandlerBadRequest(t *testing.T) {
	mockSvc := &mockService{}
	handler := New(mockSvc)

	router := gin.New()
	router.GET("/api/order/:id", handler.GetOrder)

	req, err := http.NewRequest("GET", "/api/order/", nil)
	assert.NoError(t, err)

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusNotFound, rr.Code)
}

func TestHealthCheck(t *testing.T) {
	mockSvc := &mockService{}
	handler := New(mockSvc)

	router := gin.New()
	router.GET("/api/health", handler.HealthCheck)

	req, err := http.NewRequest("GET", "/api/health", nil)
	assert.NoError(t, err)

	rr := httptest.NewRecorder()
	router.ServeHTTP(rr, req)

	assert.Equal(t, http.StatusOK, rr.Code)

	var response map[string]interface{}
	err = json.Unmarshal(rr.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "healthy", response["status"])
}