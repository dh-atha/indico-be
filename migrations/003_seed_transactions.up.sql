BEGIN;

WITH merchants AS (
    SELECT 'm-' || lpad(gs::text, 3, '0') AS merchant_id
    FROM generate_series(1, 100) gs
), days AS (
    SELECT (DATE '2025-01-01' + (gs - 1))::date AS d
    FROM generate_series(1, 100) gs
), txidx AS (
    SELECT generate_series(1, 100) AS n
)
INSERT INTO transactions (merchant_id, amount_cents, fee_cents, status, paid_at)
SELECT m.merchant_id,
       amt.amt AS amount_cents,
       GREATEST((amt.amt * 3) / 100, 30) AS fee_cents,
       'PAID' AS status,
       (d.d + (random() * INTERVAL '1 day'))::timestamptz AS paid_at
FROM merchants m
CROSS JOIN days d
CROSS JOIN txidx t
CROSS JOIN LATERAL (
    SELECT (500 + (random() * 50000)::int)::bigint AS amt
) amt;

COMMIT;