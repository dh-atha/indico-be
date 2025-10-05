package models

import "time"

type Product struct {
	ID         int64     `json:"id"`
	Name       string    `json:"name"`
	PriceCents int64     `json:"price_cents"`
	Stock      int       `json:"stock"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

type Order struct {
	ID         int64     `json:"id"`
	ProductID  int64     `json:"product_id"`
	BuyerID    string    `json:"buyer_id"`
	Quantity   int       `json:"quantity"`
	TotalCents int64     `json:"total_cents"`
	Status     string    `json:"status"`
	CreatedAt  time.Time `json:"created_at"`
}

type Transaction struct {
	ID          int64     `json:"id"`
	MerchantID  string    `json:"merchant_id"`
	AmountCents int64     `json:"amount_cents"`
	FeeCents    int64     `json:"fee_cents"`
	Status      string    `json:"status"`
	PaidAt      time.Time `json:"paid_at"`
}

type Settlement struct {
	ID          int64     `json:"id"`
	MerchantID  string    `json:"merchant_id"`
	Date        time.Time `json:"date"`
	GrossCents  int64     `json:"gross_cents"`
	FeeCents    int64     `json:"fee_cents"`
	NetCents    int64     `json:"net_cents"`
	TxnCount    int64     `json:"txn_count"`
	GeneratedAt time.Time `json:"generated_at"`
	UniqueRunID string    `json:"unique_run_id"`
}

type Job struct {
	ID              string     `json:"job_id"`
	Type            string     `json:"type"`
	Status          string     `json:"status"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
	StartedAt       *time.Time `json:"started_at,omitempty"`
	CompletedAt     *time.Time `json:"completed_at,omitempty"`
	CanceledAt      *time.Time `json:"canceled_at,omitempty"`
	CancelRequested bool       `json:"cancel_requested"`
	Total           int64      `json:"total"`
	Processed       int64      `json:"processed"`
	ResultPath      *string    `json:"result_path,omitempty"`
	Error           *string    `json:"error,omitempty"`
}
