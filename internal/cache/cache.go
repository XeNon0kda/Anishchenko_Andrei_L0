package cache

import (
	"sync"
	"order-service/internal/models"
)

type Cache struct {
	sync.RWMutex
	data map[string]*models.Order
}

func New() *Cache {
	return &Cache{
		data: make(map[string]*models.Order),
	}
}

func (c *Cache) Set(orderUID string, order *models.Order) {
	c.Lock()
	defer c.Unlock()
	c.data[orderUID] = order
}

func (c *Cache) Get(orderUID string) (*models.Order, bool) {
	c.RLock()
	defer c.RUnlock()
	order, exists := c.data[orderUID]
	return order, exists
}

func (c *Cache) Delete(orderUID string) {
	c.Lock()
	defer c.Unlock()
	delete(c.data, orderUID)
}

func (c *Cache) Size() int {
	c.RLock()
	defer c.RUnlock()
	return len(c.data)
}

func (c *Cache) GetAll() map[string]*models.Order {
	c.RLock()
	defer c.RUnlock()
	
	result := make(map[string]*models.Order)
	for k, v := range c.data {
		result[k] = v
	}
	return result
}