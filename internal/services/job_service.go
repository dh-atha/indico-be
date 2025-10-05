package services

import (
	"context"
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"

	"be/internal/repositories"
)

type JobService interface {
	StartSettlement(ctx context.Context, id string, from, to time.Time) error
	Enqueue(id string)
}

type jobService struct {
	jobs    repositories.JobRepository
	txRepo  repositories.TransactionRepository
	stRepo  repositories.SettlementRepository
	workers int

	jobQueue chan string
}

func NewJobService(j repositories.JobRepository, t repositories.TransactionRepository, s repositories.SettlementRepository, workers int) JobService {
	js := &jobService{jobs: j, txRepo: t, stRepo: s, workers: workers, jobQueue: make(chan string, 32)}
	go js.loop()
	return js
}

func (s *jobService) loop() {
	for id := range s.jobQueue {
		s.process(context.Background(), id)
	}
}

func (s *jobService) Enqueue(id string) { s.jobQueue <- id }

// StartSettlement prepares the job row and enqueues it
func (s *jobService) StartSettlement(ctx context.Context, id string, from, to time.Time) error {
	total, err := s.txRepo.CountInRange(ctx, from, to)
	if err != nil {
		return err
	}
	if err := s.jobs.Create(ctx, id, repositories.JobTypeSettlement, total, from, to); err != nil {
		return err
	}
	s.Enqueue(id)
	return nil
}

// process performs the job: stream transactions in batches, fan-out to workers to aggregate, then write CSV, upsert settlements, and update job status
func (s *jobService) process(ctx context.Context, id string) error {
	log.Println("Processing job:", id)

	if err := s.jobs.SetRunning(ctx, id); err != nil {
		return err
	}

	var err error

	start := time.Now()
	defer func() {
		if err != nil {
			log.Printf("Job %s failed after %v: %v", id, time.Since(start), err)
		}
		log.Printf("Job %s finished in %v", id, time.Since(start))
	}()

	outDir := "./tmp/settlements"
	_ = os.MkdirAll(outDir, 0o755)
	outPath := filepath.Join(outDir, fmt.Sprintf("%s.csv", id))
	f, err := os.Create(outPath)
	if err != nil {
		_ = s.jobs.SetFailed(ctx, id, err.Error())
		return err
	}
	defer f.Close()
	w := csv.NewWriter(f)
	defer w.Flush()
	_ = w.Write([]string{"merchant_id", "date", "gross", "fee", "net", "txn_count"})

	type key struct{ merchant, day string }
	agg := make(map[key]struct {
		gross, fee, net int64
		count           int64
	})
	var mu sync.Mutex

	batches := make(chan []repositories.TransactionRow, s.workers)
	var wg sync.WaitGroup
	worker := func() {
		defer wg.Done()
		log.Println("Worker started")
		for batch := range batches {
			// Check cancel request between batches
			if cancel, _ := s.jobs.IsCancelRequested(ctx, id); cancel {
				return
			}
			local := make(map[key]struct {
				gross, fee, net int64
				count           int64
			})
			for _, t := range batch {
				day := t.PaidAt.Format("2006-01-02")
				k := key{merchant: t.MerchantID, day: day}
				v := local[k]
				v.gross += t.AmountCents
				v.fee += t.FeeCents
				v.net += t.AmountCents - t.FeeCents
				v.count++
				local[k] = v
			}
			mu.Lock()
			for k, v := range local {
				a := agg[k]
				a.gross += v.gross
				a.fee += v.fee
				a.net += v.net
				a.count += v.count
				agg[k] = a
			}
			mu.Unlock()
		}
	}

	for i := 0; i < s.workers; i++ {
		wg.Add(1)
		go worker()
	}

	processed := int64(0)
	jr, _ := s.jobs.Get(ctx, id)
	var from, to time.Time
	if jr != nil && jr.FromDate.Valid {
		from = jr.FromDate.Time
	} else {
		from = time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)
	}
	if jr != nil && jr.ToDate.Valid {
		to = jr.ToDate.Time
	} else {
		to = time.Date(2025, 1, 31, 0, 0, 0, 0, time.UTC)
	}
	go func() {
		defer close(batches)
		_ = s.txRepo.StreamBatches(ctx, from, to, 10000, func(ts []repositories.TransactionRow) error {
			// Check cancel flag early
			if cancel, _ := s.jobs.IsCancelRequested(ctx, id); cancel {
				return context.Canceled
			}
			batches <- ts
			processed += int64(len(ts))
			_ = s.jobs.SetProgress(ctx, id, processed)
			return nil
		})
	}()
	wg.Wait()

	// Write CSV and upsert settlements
	for k, v := range agg {
		if err := w.Write([]string{k.merchant, k.day, fmt.Sprintf("%d", v.gross), fmt.Sprintf("%d", v.fee), fmt.Sprintf("%d", v.net), fmt.Sprintf("%d", v.count)}); err != nil {
			_ = s.jobs.SetFailed(ctx, id, err.Error())
			return err
		}
		if err := s.stRepo.Upsert(ctx, k.merchant, k.day, v.gross, v.fee, v.net, v.count, id); err != nil {
			_ = s.jobs.SetFailed(ctx, id, err.Error())
			return err
		}
	}

	if err := s.jobs.SetCompleted(ctx, id, outPath); err != nil {
		return err
	}

	return nil
}
