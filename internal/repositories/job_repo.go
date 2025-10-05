package repositories

import (
	"context"
	"database/sql"
	"time"

	"github.com/jmoiron/sqlx"
)

type JobStatus string

const (
	JobTypeSettlement = "SETTLEMENT"

	JobStatusQueued    JobStatus = "QUEUED"
	JobStatusRunning   JobStatus = "RUNNING"
	JobStatusCompleted JobStatus = "COMPLETED"
	JobStatusCanceled  JobStatus = "CANCELED"
	JobStatusFailed    JobStatus = "FAILED"
)

type JobRow struct {
	ID              string         `db:"id"`
	Type            string         `db:"type"`
	Status          JobStatus      `db:"status"`
	CreatedAt       time.Time      `db:"created_at"`
	UpdatedAt       time.Time      `db:"updated_at"`
	StartedAt       sql.NullTime   `db:"started_at"`
	CompletedAt     sql.NullTime   `db:"completed_at"`
	CanceledAt      sql.NullTime   `db:"canceled_at"`
	CancelRequested bool           `db:"cancel_requested"`
	Total           int64          `db:"total"`
	Processed       int64          `db:"processed"`
	ResultPath      sql.NullString `db:"result_path"`
	Error           sql.NullString `db:"error"`
	FromDate        sql.NullTime   `db:"from_date"`
	ToDate          sql.NullTime   `db:"to_date"`
}

type JobRepository interface {
	Create(ctx context.Context, id, typ string, total int64, from time.Time, to time.Time) error
	SetRunning(ctx context.Context, id string) error
	SetProgress(ctx context.Context, id string, processed int64) error
	SetCompleted(ctx context.Context, id string, resultPath string) error
	SetFailed(ctx context.Context, id string, msg string) error
	RequestCancel(ctx context.Context, id string) error
	Get(ctx context.Context, id string) (*JobRow, error)
	IsCancelRequested(ctx context.Context, id string) (bool, error)
}

type jobRepository struct{ db *sqlx.DB }

func NewJobRepository(db *sqlx.DB) *jobRepository { return &jobRepository{db: db} }

func (r *jobRepository) Create(ctx context.Context, id, typ string, total int64, from time.Time, to time.Time) error {
	_, err := r.db.ExecContext(ctx, `INSERT INTO jobs (id, type, status, total, from_date, to_date) VALUES ($1,$2,$3,$4,$5,$6)`, id, typ, string(JobStatusQueued), total, from, to)
	return err
}

func (r *jobRepository) SetRunning(ctx context.Context, id string) error {
	_, err := r.db.ExecContext(ctx, `UPDATE jobs SET status=$1, started_at=now(), updated_at=now() WHERE id=$2`, string(JobStatusRunning), id)
	return err
}

func (r *jobRepository) SetProgress(ctx context.Context, id string, processed int64) error {
	_, err := r.db.ExecContext(ctx, `UPDATE jobs SET processed=$1, updated_at=now() WHERE id=$2`, processed, id)
	return err
}

func (r *jobRepository) SetCompleted(ctx context.Context, id string, resultPath string) error {
	_, err := r.db.ExecContext(ctx, `UPDATE jobs SET status=$1, completed_at=now(), updated_at=now(), result_path=$2 WHERE id=$3`, string(JobStatusCompleted), resultPath, id)
	return err
}

func (r *jobRepository) SetFailed(ctx context.Context, id string, msg string) error {
	_, err := r.db.ExecContext(ctx, `UPDATE jobs SET status=$1, error=$2, updated_at=now() WHERE id=$3`, string(JobStatusFailed), msg, id)
	return err
}

func (r *jobRepository) RequestCancel(ctx context.Context, id string) error {
	_, err := r.db.ExecContext(ctx, `UPDATE jobs SET cancel_requested=true, canceled_at=now(), updated_at=now() WHERE id=$1`, id)
	return err
}

func (r *jobRepository) Get(ctx context.Context, id string) (*JobRow, error) {
	row := r.db.QueryRowxContext(ctx, `SELECT id,type,status,created_at,updated_at,started_at,completed_at,canceled_at,cancel_requested,total,processed,result_path,error,from_date,to_date FROM jobs WHERE id=$1`, id)
	var jr JobRow
	if err := row.StructScan(&jr); err != nil {
		return nil, err
	}
	return &jr, nil
}

func (r *jobRepository) IsCancelRequested(ctx context.Context, id string) (bool, error) {
	var flag bool
	err := r.db.QueryRowContext(ctx, `SELECT cancel_requested FROM jobs WHERE id=$1`, id).Scan(&flag)
	return flag, err
}
