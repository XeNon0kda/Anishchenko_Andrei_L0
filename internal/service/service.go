package service

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"order-service/internal/cache"
	"order-service/internal/models"

	"github.com/nats-io/stan.go"
)

type OrderService interface {
	ProcessMessage(data []byte) error
	GetOrder(orderUID string) (*models.Order, error)
	GetCacheSize() int
	RestoreCache() error
}

type orderService struct {
	db       *sql.DB
	cache    *cache.Cache
	stanConn stan.Conn
}

func New(db *sql.DB, cache *cache.Cache, stanConn stan.Conn) OrderService {
	return &orderService{
		db:       db,
		cache:    cache,
		stanConn: stanConn,
	}
}

func (s *orderService) ProcessMessage(data []byte) error {
	if s.stanConn == nil {
		return fmt.Errorf("NATS connection not initialized")
	}
	
	var order models.Order
	if err := json.Unmarshal(data, &order); err != nil {
		return fmt.Errorf("invalid JSON: %v", err)
	}

	if order.OrderUID == "" {
		return fmt.Errorf("order_uid is required")
	}

	if err := s.saveOrder(&order); err != nil {
		return fmt.Errorf("failed to save order: %v", err)
	}

	s.cache.Set(order.OrderUID, &order)
	log.Printf("Order %s processed successfully", order.OrderUID)
	return nil
}

func (s *orderService) GetOrder(orderUID string) (*models.Order, error) {
	order, exists := s.cache.Get(orderUID)
	if !exists {
		return nil, fmt.Errorf("order not found")
	}
	return order, nil
}

func (s *orderService) GetCacheSize() int {
	return s.cache.Size()
}

func (s *orderService) saveOrder(order *models.Order) error {
	ctx := context.Background()
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	_, err = tx.ExecContext(ctx, `
		INSERT INTO orders (order_uid, track_number, entry, locale, internal_signature, 
			customer_id, delivery_service, shardkey, sm_id, date_created, oof_shard)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		ON CONFLICT (order_uid) DO UPDATE SET
			track_number = EXCLUDED.track_number,
			entry = EXCLUDED.entry,
			locale = EXCLUDED.locale,
			internal_signature = EXCLUDED.internal_signature,
			customer_id = EXCLUDED.customer_id,
			delivery_service = EXCLUDED.delivery_service,
			shardkey = EXCLUDED.shardkey,
			sm_id = EXCLUDED.sm_id,
			date_created = EXCLUDED.date_created,
			oof_shard = EXCLUDED.oof_shard`,
		order.OrderUID, order.TrackNumber, order.Entry, order.Locale, order.InternalSignature,
		order.CustomerID, order.DeliveryService, order.Shardkey, order.SmID, order.DateCreated, order.OofShard)
	if err != nil {
		return err
	}

	_, err = tx.ExecContext(ctx, `
		INSERT INTO delivery (order_uid, name, phone, zip, city, address, region, email)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		ON CONFLICT (order_uid) DO UPDATE SET
			name = EXCLUDED.name,
			phone = EXCLUDED.phone,
			zip = EXCLUDED.zip,
			city = EXCLUDED.city,
			address = EXCLUDED.address,
			region = EXCLUDED.region,
			email = EXCLUDED.email`,
		order.OrderUID, order.Delivery.Name, order.Delivery.Phone, order.Delivery.Zip,
		order.Delivery.City, order.Delivery.Address, order.Delivery.Region, order.Delivery.Email)
	if err != nil {
		return err
	}

	_, err = tx.ExecContext(ctx, `
		INSERT INTO payment (transaction, order_uid, request_id, currency, provider, 
			amount, payment_dt, bank, delivery_cost, goods_total, custom_fee)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		ON CONFLICT (transaction) DO UPDATE SET
			order_uid = EXCLUDED.order_uid,
			request_id = EXCLUDED.request_id,
			currency = EXCLUDED.currency,
			provider = EXCLUDED.provider,
			amount = EXCLUDED.amount,
			payment_dt = EXCLUDED.payment_dt,
			bank = EXCLUDED.bank,
			delivery_cost = EXCLUDED.delivery_cost,
			goods_total = EXCLUDED.goods_total,
			custom_fee = EXCLUDED.custom_fee`,
		order.Payment.Transaction, order.OrderUID, order.Payment.RequestID, order.Payment.Currency,
		order.Payment.Provider, order.Payment.Amount, order.Payment.PaymentDt, order.Payment.Bank,
		order.Payment.DeliveryCost, order.Payment.GoodsTotal, order.Payment.CustomFee)
	if err != nil {
		return err
	}

	_, err = tx.ExecContext(ctx, "DELETE FROM items WHERE order_uid = $1", order.OrderUID)
	if err != nil {
		return err
	}

	for _, item := range order.Items {
		_, err = tx.ExecContext(ctx, `
			INSERT INTO items (order_uid, chrt_id, track_number, price, rid, name, 
				sale, size, total_price, nm_id, brand, status)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)`,
			order.OrderUID, item.ChrtID, item.TrackNumber, item.Price, item.Rid, item.Name,
			item.Sale, item.Size, item.TotalPrice, item.NmID, item.Brand, item.Status)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (s *orderService) RestoreCache() error {
	rows, err := s.db.Query(`
		SELECT o.order_uid, o.track_number, o.entry, o.locale, o.internal_signature,
			o.customer_id, o.delivery_service, o.shardkey, o.sm_id, o.date_created, o.oof_shard,
			d.name, d.phone, d.zip, d.city, d.address, d.region, d.email,
			p.transaction, p.request_id, p.currency, p.provider, p.amount, p.payment_dt,
			p.bank, p.delivery_cost, p.goods_total, p.custom_fee
		FROM orders o
		LEFT JOIN delivery d ON o.order_uid = d.order_uid
		LEFT JOIN payment p ON o.order_uid = p.order_uid
	`)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var order models.Order
		var delivery models.Delivery
		var payment models.Payment

		err := rows.Scan(
			&order.OrderUID, &order.TrackNumber, &order.Entry, &order.Locale, &order.InternalSignature,
			&order.CustomerID, &order.DeliveryService, &order.Shardkey, &order.SmID, &order.DateCreated, &order.OofShard,
			&delivery.Name, &delivery.Phone, &delivery.Zip, &delivery.City, &delivery.Address, &delivery.Region, &delivery.Email,
			&payment.Transaction, &payment.RequestID, &payment.Currency, &payment.Provider, &payment.Amount, &payment.PaymentDt,
			&payment.Bank, &payment.DeliveryCost, &payment.GoodsTotal, &payment.CustomFee,
		)
		if err != nil {
			return err
		}

		order.Delivery = delivery
		order.Payment = payment

		itemRows, err := s.db.Query(`
			SELECT chrt_id, track_number, price, rid, name, sale, size, 
				total_price, nm_id, brand, status
			FROM items WHERE order_uid = $1`, order.OrderUID)
		if err != nil {
			return err
		}

		var items []models.Item
		for itemRows.Next() {
			var item models.Item
			err := itemRows.Scan(
				&item.ChrtID, &item.TrackNumber, &item.Price, &item.Rid, &item.Name, &item.Sale,
				&item.Size, &item.TotalPrice, &item.NmID, &item.Brand, &item.Status,
			)
			if err != nil {
				itemRows.Close()
				return err
			}
			items = append(items, item)
		}
		itemRows.Close()

		order.Items = items
		s.cache.Set(order.OrderUID, &order)
	}

	log.Printf("Cache restored with %d orders", s.cache.Size())
	return nil
}