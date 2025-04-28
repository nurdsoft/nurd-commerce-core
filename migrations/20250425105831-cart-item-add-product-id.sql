-- +migrate Up
ALTER TABLE order_items
    ADD COLUMN product_id UUID NULL REFERENCES products (id);
-- +migrate Down
ALTER TABLE order_items
    DROP COLUMN product_id;