-- +migrate Up
ALTER TABLE orders
ALTER COLUMN shipping_rate DROP NOT NULL;
-- +migrate Down
ALTER TABLE orders
ALTER COLUMN shipping_rate SET NOT NULL;
