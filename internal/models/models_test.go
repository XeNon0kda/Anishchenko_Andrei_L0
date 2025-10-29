package models

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestJSONValidation(t *testing.T) {
	testOrder := Order{
		OrderUID:    "test-order-123",
		TrackNumber: "WBILMTESTTRACK",
		Entry:       "WBIL",
		Locale:      "en",
		CustomerID:  "test",
		Delivery: Delivery{
			Name: "Test Testov",
		},
		Payment: Payment{
			Transaction: "test-transaction",
		},
		Items: []Item{},
		DateCreated: time.Now(),
	}

	data, err := json.Marshal(testOrder)
	assert.NoError(t, err)

	var order Order
	err = json.Unmarshal(data, &order)
	assert.NoError(t, err)

	assert.Equal(t, testOrder.OrderUID, order.OrderUID)
	assert.Equal(t, testOrder.TrackNumber, order.TrackNumber)
	assert.Equal(t, testOrder.Entry, order.Entry)
}

func TestJSONInvalid(t *testing.T) {
	invalidJSON := []byte(`{invalid json}`)
	
	var order Order
	err := json.Unmarshal(invalidJSON, &order)
	assert.Error(t, err)
}