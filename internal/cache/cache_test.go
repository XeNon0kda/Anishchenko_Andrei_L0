package cache

import (
	"testing"

	"order-service/internal/models"

	"github.com/stretchr/testify/assert"
)

func TestCacheOperations(t *testing.T) {
	cache := New()

	testOrder := &models.Order{
		OrderUID: "cache-test-123",
	}

	cache.Set(testOrder.OrderUID, testOrder)

	order, exists := cache.Get(testOrder.OrderUID)
	assert.True(t, exists)
	assert.Equal(t, testOrder.OrderUID, order.OrderUID)

	_, exists = cache.Get("nonexistent")
	assert.False(t, exists)

	cache.Delete(testOrder.OrderUID)
	_, exists = cache.Get(testOrder.OrderUID)
	assert.False(t, exists)
}

func TestCacheConcurrent(t *testing.T) {
	cache := New()

	cache.Set("order1", &models.Order{OrderUID: "order1"})
	cache.Set("order2", &models.Order{OrderUID: "order2"})

	order1, exists1 := cache.Get("order1")
	order2, exists2 := cache.Get("order2")

	assert.True(t, exists1)
	assert.True(t, exists2)
	assert.Equal(t, "order1", order1.OrderUID)
	assert.Equal(t, "order2", order2.OrderUID)
}