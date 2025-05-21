-- +migrate Up
ALTER TYPE order_status ADD VALUE 'processing';
ALTER TYPE order_status ADD VALUE 'packed';
-- +migrate Down
-- There is no ALTER TYPE DELETE VALUE in Postgres. You can only add new values.