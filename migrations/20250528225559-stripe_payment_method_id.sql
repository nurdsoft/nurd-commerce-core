-- +migrate Up
ALTER TABLE orders
ADD COLUMN IF NOT EXISTS stripe_payment_method_id text;
-- +migrate Down
ALTER TABLE orders
DROP COLUMN IF EXISTS stripe_payment_method_id;
