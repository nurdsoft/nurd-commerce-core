-- +migrate Up
-- Add shipping and fulfillment fields to order items
ALTER TABLE order_items
ADD COLUMN shipping_rate_id UUID REFERENCES cart_shipping_rates(id),
ADD COLUMN shipping_rate NUMERIC(10, 2),
ADD COLUMN shipping_carrier_name TEXT,
ADD COLUMN shipping_carrier_code TEXT,
ADD COLUMN shipping_service_type TEXT,
ADD COLUMN shipping_service_code TEXT,
ADD COLUMN estimated_delivery_date TIMESTAMPTZ,
ADD COLUMN business_days_in_transit TEXT,
ADD COLUMN tracking_number TEXT,
ADD COLUMN tracking_url TEXT,
ADD COLUMN shipment_date TIMESTAMPTZ,
ADD COLUMN freight_charge NUMERIC(10, 2),
ADD COLUMN amount_due NUMERIC(10, 2),
ADD COLUMN fulfillment_message TEXT,
ADD COLUMN fulfillment_metadata JSONB;

-- +migrate Down
-- Remove shipping and fulfillment fields from order items
ALTER TABLE order_items
DROP COLUMN shipping_rate_id,
DROP COLUMN shipping_rate,
DROP COLUMN shipping_carrier_name,
DROP COLUMN shipping_carrier_code,
DROP COLUMN shipping_service_type,
DROP COLUMN shipping_service_code,
DROP COLUMN estimated_delivery_date,
DROP COLUMN business_days_in_transit,
DROP COLUMN tracking_number,
DROP COLUMN tracking_url,
DROP COLUMN shipment_date,
DROP COLUMN freight_charge,
DROP COLUMN amount_due,
DROP COLUMN fulfillment_message,
DROP COLUMN fulfillment_metadata;