-- +migrate Up
ALTER TABLE customers RENAME COLUMN stripe_id TO external_customer_id;

ALTER TABLE orders RENAME COLUMN stripe_payment_intent_id TO external_payment_id; 

-- +migrate Down
ALTER TABLE customers RENAME COLUMN external_customer_id TO stripe_id;

ALTER TABLE orders RENAME COLUMN external_payment_id TO stripe_payment_intent_id;