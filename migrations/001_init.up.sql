BEGIN;

-- Products table
CREATE TABLE IF NOT EXISTS products (
    id BIGSERIAL PRIMARY KEY,
    name TEXT NOT NULL,
    price_cents BIGINT NOT NULL DEFAULT 0 CHECK (price_cents >= 0),
    stock INTEGER NOT NULL DEFAULT 0 CHECK (stock >= 0),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_products_stock ON products(stock);

-- Orders table
CREATE TABLE IF NOT EXISTS orders (
    id BIGSERIAL PRIMARY KEY,
    product_id BIGINT NOT NULL REFERENCES products(id),
    buyer_id TEXT NOT NULL,
    quantity INTEGER NOT NULL CHECK (quantity > 0),
    total_cents BIGINT NOT NULL CHECK (total_cents >= 0),
    status TEXT NOT NULL DEFAULT 'CREATED',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS idx_orders_product_id ON orders(product_id);
CREATE INDEX IF NOT EXISTS idx_orders_created_at ON orders(created_at);

-- Transactions table
CREATE TABLE IF NOT EXISTS transactions (
    id BIGSERIAL PRIMARY KEY,
    merchant_id TEXT NOT NULL,
    amount_cents BIGINT NOT NULL CHECK (amount_cents >= 0),
    fee_cents BIGINT NOT NULL CHECK (fee_cents >= 0),
    status TEXT NOT NULL,
    paid_at TIMESTAMPTZ NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_transactions_merchant_date ON transactions(merchant_id, paid_at);
CREATE INDEX IF NOT EXISTS idx_transactions_status ON transactions(status);

-- Settlements table
CREATE TABLE IF NOT EXISTS settlements (
    id BIGSERIAL PRIMARY KEY,
    merchant_id TEXT NOT NULL,
    date DATE NOT NULL,
    gross_cents BIGINT NOT NULL CHECK (gross_cents >= 0),
    fee_cents BIGINT NOT NULL CHECK (fee_cents >= 0),
    net_cents BIGINT NOT NULL CHECK (net_cents >= 0),
    txn_count BIGINT NOT NULL CHECK (txn_count >= 0),
    generated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    unique_run_id TEXT NOT NULL,
    UNIQUE (merchant_id, date)
);
CREATE INDEX IF NOT EXISTS idx_settlements_merchant_date ON settlements(merchant_id, date);

-- Jobs table
CREATE TABLE IF NOT EXISTS jobs (
    id TEXT PRIMARY KEY,
    type TEXT NOT NULL,
    status TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    started_at TIMESTAMPTZ,
    completed_at TIMESTAMPTZ,
    canceled_at TIMESTAMPTZ,
    cancel_requested BOOLEAN NOT NULL DEFAULT FALSE,
    total BIGINT NOT NULL DEFAULT 0,
    processed BIGINT NOT NULL DEFAULT 0,
    result_path TEXT,
    error TEXT
);
CREATE INDEX IF NOT EXISTS idx_jobs_status ON jobs(status);

COMMIT;
