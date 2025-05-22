-- +migrate Up
ALTER TABLE orders ADD COLUMN fulfillment_metadata JSONB DEFAULT '{}'::JSONB;
-- +migrate Down
ALTER TABLE orders DROP COLUMN fulfillment_metadata;
