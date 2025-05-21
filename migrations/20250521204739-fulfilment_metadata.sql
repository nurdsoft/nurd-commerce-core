-- +migrate Up
ALTER TABLE orders ADD COLUMN fulfilment_metadata JSONB DEFAULT '{}'::JSONB;
-- +migrate Down
ALTER TABLE orders DROP COLUMN fulfilment_metadata;
