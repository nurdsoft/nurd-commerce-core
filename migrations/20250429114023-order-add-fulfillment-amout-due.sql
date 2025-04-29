-- +migrate Up
ALTER TABLE orders ADD COLUMN fulfillment_amount_due DECIMAL NULL;
-- +migrate Down
ALTER TABLE orders DROP COLUMN fulfillment_amount_due;