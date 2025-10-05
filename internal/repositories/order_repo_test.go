package repositories

import (
	"context"
	"log"
	"os"
	"sync"
	"testing"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

func setupTestDB(t *testing.T) *sqlx.DB {
	t.Helper()
	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		dsn = "postgres://postgres:password@localhost:5432/appdb?sslmode=disable"
	}
	db, err := sqlx.Open("postgres", dsn)
	if err != nil {
		t.Skipf("skipping: cannot open db: %v", err)
	}
	if err := db.Ping(); err != nil {
		t.Skipf("skipping: cannot ping db: %v", err)
	}
	db.SetMaxOpenConns(50)
	db.SetMaxIdleConns(10)
	db.SetConnMaxLifetime(0)
	return db
}

// TestNoOversell launches 500 concurrent orders for a product with 100 stock.
// Expect exactly 100 successes and 400 OUT_OF_STOCK.
func TestNoOversell(t *testing.T) {
	db := setupTestDB(t)
	repo := NewOrderRepository(db)
	ctx := context.Background()
	var productID int64
	if err := db.QueryRowxContext(ctx, `INSERT INTO products (name, price_cents, stock) VALUES ('Test',100,100)
        RETURNING ID`).Scan(&productID); err != nil {
		t.Fatal(err)
	}

	// reset orders table
	if _, err := db.Exec(`DELETE FROM orders where buyer_id LIKE 'userTest-%'`); err != nil {
		t.Fatal(err)
	}

	const buyers = 500
	var wg sync.WaitGroup
	wg.Add(buyers)
	var okCount, outCount, otherErr int64
	var mu sync.Mutex
	for i := 0; i < buyers; i++ {
		go func(i int) {
			defer wg.Done()
			_, err := repo.CreateOrderWithStock(ctx, productID, 1, "userTest-")
			mu.Lock()
			defer mu.Unlock()
			switch err {
			case nil:
				okCount++
			case ErrOutOfStock:
				outCount++
			default:
				log.Println("Order error:", err)
				otherErr++
			}
		}(i)
	}
	wg.Wait()

	if okCount != 100 || outCount != 400 || otherErr != 0 {
		t.Fatalf("expected 100 successful orders, got %d (out=%d, other=%d)", okCount, outCount, otherErr)
	}
}
