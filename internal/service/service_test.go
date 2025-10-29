package service

import (
	"database/sql"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"order-service/internal/cache"
	"order-service/internal/models"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/nats-io/stan.go"
	"github.com/stretchr/testify/assert"
)

func TestProcessMessage_ValidJSON(t *testing.T) {
	db, mock, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	cache := cache.New()
	stanConn := &mockStanConn{}

	service := New(db, cache, stanConn)

	order := models.Order{
		OrderUID:    "test-123",
		TrackNumber: "TRACK123",
		Entry:       "WBIL",
		Locale:      "en",
		CustomerID:  "test",
		DateCreated: time.Now(),
		Delivery: models.Delivery{
			Name:    "Test",
			Phone:   "+123456789",
			Zip:     "123456",
			City:    "Moscow",
			Address: "Test Address",
			Region:  "Moscow",
			Email:   "test@test.com",
		},
		Payment: models.Payment{
			Transaction: "test-transaction",
			Currency:    "USD",
			Provider:    "wbpay",
			Amount:      1000,
			PaymentDt:   1637907727,
			Bank:        "bank",
			DeliveryCost: 500,
			GoodsTotal:  500,
			CustomFee:   0,
		},
		Items: []models.Item{},
	}

	data, err := json.Marshal(order)
	assert.NoError(t, err)

	mock.ExpectBegin()
	mock.ExpectExec("INSERT INTO orders").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec("INSERT INTO delivery").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec("INSERT INTO payment").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec("DELETE FROM items").WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectCommit()

	err = service.ProcessMessage(data)
	assert.NoError(t, err)

	cachedOrder, exists := cache.Get(order.OrderUID)
	assert.True(t, exists)
	assert.Equal(t, order.OrderUID, cachedOrder.OrderUID)
}

func TestProcessMessage_InvalidJSON(t *testing.T) {
	db, _, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	cache := cache.New()
	stanConn := &mockStanConn{}

	service := New(db, cache, stanConn)

	invalidJSON := []byte(`{invalid json}`)
	err = service.ProcessMessage(invalidJSON)
	assert.Error(t, err)
}

func TestProcessMessage_MissingOrderUID(t *testing.T) {
	db, _, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	cache := cache.New()
	stanConn := &mockStanConn{}

	service := New(db, cache, stanConn)

	order := map[string]interface{}{
		"track_number": "TRACK123",
	}
	data, _ := json.Marshal(order)

	err = service.ProcessMessage(data)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "order_uid is required")
}

func TestGetOrder(t *testing.T) {
	db, _, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	cache := cache.New()
	stanConn := &mockStanConn{}

	service := New(db, cache, stanConn)

	testOrder := &models.Order{
		OrderUID:    "test-123",
		TrackNumber: "TRACK123",
	}

	cache.Set(testOrder.OrderUID, testOrder)

	order, err := service.GetOrder("test-123")
	assert.NoError(t, err)
	assert.Equal(t, testOrder.OrderUID, order.OrderUID)

	_, err = service.GetOrder("nonexistent")
	assert.Error(t, err)
}

func TestGetCacheSize(t *testing.T) {
	db, _, err := sqlmock.New()
	assert.NoError(t, err)
	defer db.Close()

	cache := cache.New()
	stanConn := &mockStanConn{}

	service := New(db, cache, stanConn)

	assert.Equal(t, 0, service.GetCacheSize())

	cache.Set("test1", &models.Order{OrderUID: "test1"})
	cache.Set("test2", &models.Order{OrderUID: "test2"})

	assert.Equal(t, 2, service.GetCacheSize())
}

type mockStanConn struct{}

func (m *mockStanConn) Publish(subject string, data []byte) error {
	return nil
}

func (m *mockStanConn) PublishAsync(subject string, data []byte, ah stan.AckHandler) (string, error) {
	return "", nil
}

func (m *mockStanConn) Subscribe(subject string, cb stan.MsgHandler, opts ...stan.SubscriptionOption) (stan.Subscription, error) {
	return &mockSubscription{}, nil
}

func (m *mockStanConn) QueueSubscribe(subject, qgroup string, cb stan.MsgHandler, opts ...stan.SubscriptionOption) (stan.Subscription, error) {
	return &mockSubscription{}, nil
}

func (m *mockStanConn) Close() error {
	return nil
}

func (m *mockStanConn) NatsConn() stan.Conn {
	return nil
}

type mockSubscription struct{}

func (m *mockSubscription) Unsubscribe() error {
	return nil
}

func (m *mockSubscription) Close() error {
	return nil
}