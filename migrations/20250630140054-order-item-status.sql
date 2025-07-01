-- +migrate Up

CREATE TYPE order_item_status AS ENUM (
    'pending',
    'processing',
    'shipped',
    'fulfillment_failed',
    'delivered',
    'cancelled',
    'return_requested',
    'returned',
    'refunded'
);

ALTER TABLE order_items
ADD COLUMN IF NOT EXISTS status order_item_status NOT NULL DEFAULT 'pending';
-- +migrate Down
ALTER TABLE order_items
DROP COLUMN IF EXISTS status;

DROP TYPE IF EXISTS order_item_status;
