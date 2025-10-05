package repositories

import (
	"context"
	"log"
	"time"

	"github.com/jmoiron/sqlx"
)

type TransactionRow struct {
	ID          int64     `db:"id"`
	MerchantID  string    `db:"merchant_id"`
	AmountCents int64     `db:"amount_cents"`
	FeeCents    int64     `db:"fee_cents"`
	Status      string    `db:"status"`
	PaidAt      time.Time `db:"paid_at"`
}

type TransactionRepository interface {
	CountInRange(ctx context.Context, from, to time.Time) (int64, error)
	StreamBatches(ctx context.Context, from, to time.Time, batchSize int, fn func([]TransactionRow) error) error
}

type transactionRepository struct{ db *sqlx.DB }

func NewTransactionRepository(db *sqlx.DB) TransactionRepository {
	return &transactionRepository{db: db}
}

// Count in date range (inclusive)
func (r *transactionRepository) CountInRange(ctx context.Context, from, to time.Time) (int64, error) {
	var cnt int64
	err := r.db.QueryRowContext(ctx, `SELECT COUNT(1) FROM transactions WHERE paid_at >= $1 AND paid_at < $2 AND status='PAID'`, from, to.Add(24*time.Hour)).Scan(&cnt)
	return cnt, err
}

// StreamBatches yields transactions in batches via callback to avoid loading all in memory
func (r *transactionRepository) StreamBatches(ctx context.Context, from, to time.Time, batchSize int, fn func([]TransactionRow) error) error {
	var lastID int64 = 0
	end := to.Add(24 * time.Hour)
	for {
		log.Printf("Fetching transaction row from id: %d limit: %d\n", lastID, batchSize)
		log.Printf("Streaming from %v to %v (end=%v)\n", from, to, end)
		rows, err := r.db.QueryxContext(ctx, `SELECT id, merchant_id, amount_cents, fee_cents, status, paid_at 
            FROM transactions 
            WHERE paid_at >= $1 AND paid_at < $2 AND status='PAID' AND id > $3 
            ORDER BY id ASC 
            LIMIT $4`, from, end, lastID, batchSize)
		if err != nil {
			log.Println("Error querying transactions:", err)
			return err
		}
		batch := make([]TransactionRow, 0, batchSize)
		for rows.Next() {
			var t TransactionRow
			if err := rows.StructScan(&t); err != nil {
				rows.Close()
				log.Println("Error scanning transaction row:", err)
				return err
			}
			batch = append(batch, t)
		}
		rows.Close()
		if len(batch) == 0 {
			log.Println("No more transactions to process")
			return nil
		}
		lastID = batch[len(batch)-1].ID
		if err := fn(batch); err != nil {
			log.Println("Error processing batch:", err)
			return err
		}
		if len(batch) < batchSize {
			log.Println("Processed all available transactions")
			return nil
		}
	}
}
