-- +migrate Up
ALTER TABLE order_items
ADD COLUMN IF NOT EXISTS stripe_refund_id text NULL;
ALTER TABLE order_items
ADD COLUMN IF NOT EXISTS stripe_refund_amount numeric(10, 2) NULL;
ALTER TABLE order_items
ADD COLUMN IF NOT EXISTS stripe_refund_created_at text NULL;

ALTER TABLE orders
ADD COLUMN IF NOT EXISTS stripe_refund_total numeric(10, 2) NULL;
-- +migrate Down
ALTER TABLE order_items
DROP COLUMN IF EXISTS stripe_refund_id;
ALTER TABLE order_items
DROP COLUMN IF EXISTS stripe_refund_amount;
ALTER TABLE order_items
DROP COLUMN IF EXISTS stripe_refund_created_at;

ALTER TABLE orders
DROP COLUMN IF EXISTS stripe_refund_total;
