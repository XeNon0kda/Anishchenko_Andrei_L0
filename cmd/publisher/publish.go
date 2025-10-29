package main

import (
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/nats-io/stan.go"
)

func main() {
	sc, err := stan.Connect("my-cluster", "test-publisher")
	if err != nil {
		log.Fatal(err)
	}
	defer sc.Close()

	order := map[string]interface{}{
		"order_uid":          "RWBLABS",
		"track_number":       "TESTTRACK",
		"entry":              "WB",
		"delivery": map[string]interface{}{
			"name":    "Wild Berry",
			"phone":   "+78005553535",
			"zip":     "1234567",
			"city":    "Novosibirsk",
			"address": "Kamenskaya 52/1",
			"region":  "Novosib",
			"email":   "wb-nsuem@gmail.com",
		},
		"payment": map[string]interface{}{
			"transaction":   "123ab764de935f0a228",
			"request_id":    "",
			"currency":      "RUB",
			"provider":      "wbpay",
			"amount":        1337,
			"payment_dt":    1637907727,
			"bank":          "VTB",
			"delivery_cost": 1000,
			"goods_total":   337,
			"custom_fee":    0,
		},
		"items": []map[string]interface{}{
			{
				"chrt_id":      9034930,
				"track_number": "WBILMTESTTRACK",
				"price":        453,
				"rid":          "ab4219087a764ae0btest",
				"name":         "Mascaras",
				"sale":         30,
				"size":         "0",
				"total_price":  317,
				"nm_id":        2389212,
				"brand":        "Vivienne Sabo",
				"status":       202,
			},
		},
		"locale":             "en",
		"internal_signature": "",
		"customer_id":        "test",
		"delivery_service":   "meest",
		"shardkey":           "9",
		"sm_id":              99,
		"date_created":       time.Now().Format(time.RFC3339),
		"oof_shard":          "1",
	}

	data, _ := json.Marshal(order)
	
	err = sc.Publish("orders", data)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Test message sent successfully!")
	fmt.Println("Order UID: test-order-123")
}