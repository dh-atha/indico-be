package repositories

import (
	"context"

	"github.com/jmoiron/sqlx"
)

type SettlementRepository interface {
	Upsert(ctx context.Context, merchantID string, date string, gross, fee, net int64, txnCount int64, runID string) error
}

type settlementRepository struct{ db *sqlx.DB }

func NewSettlementRepository(db *sqlx.DB) SettlementRepository { return &settlementRepository{db: db} }

// Upsert merchant/day row
func (r *settlementRepository) Upsert(ctx context.Context, merchantID string, date string, gross, fee, net int64, txnCount int64, runID string) error {
	_, err := r.db.ExecContext(ctx, `INSERT INTO settlements (merchant_id, date, gross_cents, fee_cents, net_cents, txn_count, unique_run_id) 
        VALUES ($1,$2,$3,$4,$5,$6,$7)
        ON CONFLICT (merchant_id, date) DO UPDATE SET 
           gross_cents=EXCLUDED.gross_cents,
           fee_cents=EXCLUDED.fee_cents,
           net_cents=EXCLUDED.net_cents,
           txn_count=EXCLUDED.txn_count,
           unique_run_id=EXCLUDED.unique_run_id,
           generated_at=now()`, merchantID, date, gross, fee, net, txnCount, runID)
	return err
}
