-- +migrate Up
-- Add shipping rate ID to cart items
ALTER TABLE cart_items
ADD COLUMN shipping_rate_id UUID REFERENCES cart_shipping_rates(id);

-- +migrate Down
-- Remove shipping rate ID from cart items
ALTER TABLE cart_items
DROP COLUMN shipping_rate_id;