package models

type Order struct {
	OrderUID    string   `json:"order_uid"`
	TrackNumber string   `json:"track_number"`
	Entry       string   `json:"entry"`
	Delivery    Delivery `json:"delivery"`
	Payment     Payment  `json:"payment"`
	Items       []Item   `json:"items"`
	Locale      string   `json:"locale"`
	InternalSig string   `json:"internal_signature"`
	CustomerID  string   `json:"customer_id"`
	DeliverySvc string   `json:"delivery_service"`
	ShardKey    string   `json:"shardkey"`
	SmID        int      `json:"sm_id"`
	DateCreated string   `json:"date_created"`
	OofShard    string   `json:"oof_shard"`
}
