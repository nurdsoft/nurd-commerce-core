-- +migrate Up
ALTER TABLE customers ADD COLUMN IF NOT EXISTS authorizenet_id TEXT UNIQUE;

ALTER TABLE orders ADD COLUMN IF NOT EXISTS authorizenet_payment_id TEXT;

-- +migrate Down
ALTER TABLE customers DROP COLUMN IF EXISTS authorizenet_id;

ALTER TABLE orders DROP COLUMN IF EXISTS authorizenet_payment_id;