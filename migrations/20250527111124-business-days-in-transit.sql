-- +migrate Up
ALTER TABLE cart_shipping_rates
    ADD COLUMN business_days_in_transit TEXT;
ALTER TABLE orders
    ADD COLUMN shipping_business_days_in_transit TEXT;
-- +migrate Down
ALTER TABLE cart_shipping_rates DROP COLUMN business_days_in_transit;
ALTER TABLE orders DROP COLUMN shipping_business_days_in_transit;