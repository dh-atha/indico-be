BEGIN;

INSERT INTO products (name, price_cents, stock)
VALUES ('Limited Edition Widget', 1999, 100)
ON CONFLICT (id) DO UPDATE SET
    name = EXCLUDED.name,
    price_cents = EXCLUDED.price_cents,
    stock = EXCLUDED.stock,
    updated_at = now();

COMMIT;