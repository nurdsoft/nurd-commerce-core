-- +migrate Up
ALTER TABLE orders ADD COLUMN IF NOT EXISTS fulfillment_tracking_url TEXT;

-- +migrate Down
ALTER TABLE orders DROP COLUMN IF EXISTS fulfillment_tracking_url;
