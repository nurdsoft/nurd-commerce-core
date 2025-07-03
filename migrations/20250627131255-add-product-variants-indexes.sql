-- +migrate Up

-- Enable the pg_trgm extension for trigram-based text search indexes
CREATE EXTENSION IF NOT EXISTS pg_trgm;

-- GIN index for text search (name ILIKE and description ILIKE)
CREATE INDEX idx_product_variants_name_description_search 
ON product_variants USING GIN (name gin_trgm_ops, description gin_trgm_ops);

-- B-tree index for price range queries (price >= and price <=)
CREATE INDEX idx_product_variants_price 
ON product_variants (price);

-- GIN index for JSON attributes queries (attributes->>'key' = 'value')
CREATE INDEX idx_product_variants_attributes 
ON product_variants USING GIN (attributes);

-- B-tree index for default sorting (created_at DESC)
CREATE INDEX idx_product_variants_created_at_desc 
ON product_variants (created_at DESC);

-- +migrate Down

DROP INDEX IF EXISTS idx_product_variants_name_description_search;
DROP INDEX IF EXISTS idx_product_variants_price;
DROP INDEX IF EXISTS idx_product_variants_attributes;
DROP INDEX IF EXISTS idx_product_variants_created_at_desc; 