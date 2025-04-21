-- +migrate Up

CREATE TABLE customers
(
    id UUID NOT NULL PRIMARY KEY,
    email VARCHAR(255) NOT NULL UNIQUE,
    first_name VARCHAR(255) NOT NULL,
    last_name VARCHAR(255),
    phone_number VARCHAR(50),
    salesforce_id TEXT UNIQUE,
    stripe_id TEXT UNIQUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ
);

CREATE TABLE addresses
(
    id UUID NOT NULL PRIMARY KEY,
    customer_id UUID NOT NULL REFERENCES customers (id),
    full_name VARCHAR(255) NOT NULL,
    address VARCHAR(255) NOT NULL,
    apartment VARCHAR(255),
    city VARCHAR(255),
    phone_number VARCHAR(50),
    state_code VARCHAR(255) NOT NULL,
    country_code VARCHAR(255) NOT NULL,
    postal_code VARCHAR(255) NOT NULL,
    is_default BOOLEAN NOT NULL DEFAULT false,
    salesforce_id TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ
);

CREATE TABLE products
(
    id UUID NOT NULL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    image_url TEXT,
    attributes JSONB,
    salesforce_id TEXT,
    salesforce_pricebook_entry_id TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ,
    CONSTRAINT products_salesforce_unique_key
    UNIQUE (id, salesforce_id, salesforce_pricebook_entry_id)
);

CREATE TABLE product_variants
(
    id UUID NOT NULL PRIMARY KEY,
    product_id UUID NOT NULL REFERENCES products (id),
    sku VARCHAR(255) NOT NULL UNIQUE,
    image_url TEXT,
    description TEXT,
    name VARCHAR(255) NOT NULL,
    price NUMERIC(10, 2) NOT NULL,
    currency VARCHAR(3) NOT NULL,
    length NUMERIC(10, 2),
    width NUMERIC(10, 2),
    height NUMERIC(10, 2),
    weight NUMERIC(10, 3),
    attributes JSONB,
    stripe_tax_code TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ
);

CREATE TABLE wishlist_items
(
    id UUID NOT NULL PRIMARY KEY,
    customer_id UUID NOT NULL REFERENCES customers (id),
    product_id UUID NOT NULL REFERENCES products (id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    CONSTRAINT wishlist_items_unique_key UNIQUE (customer_id, product_id)
);

CREATE TYPE cart_status AS ENUM ('active', 'purchased', 'cleared');

CREATE TABLE carts
(
    id UUID NOT NULL PRIMARY KEY,
    customer_id UUID NOT NULL REFERENCES customers (id),
    status CART_STATUS NOT NULL,
    tax_currency VARCHAR(3),
    tax_amount NUMERIC(10, 2),
    tax_breakdown JSONB,
    shipping_rate_id UUID,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ
);

CREATE TABLE cart_items
(
    id UUID NOT NULL PRIMARY KEY,
    cart_id UUID NOT NULL REFERENCES carts (id),
    product_variant_id UUID NOT NULL REFERENCES product_variants (id),
    quantity INTEGER NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ
);

CREATE TABLE cart_shipping_rates
(
    id UUID NOT NULL PRIMARY KEY,
    cart_id UUID NOT NULL REFERENCES carts (id),
    address_id UUID NOT NULL REFERENCES addresses (id),
    amount NUMERIC(10, 2) NOT NULL,
    currency VARCHAR(3) NOT NULL,
    carrier_name TEXT,
    carrier_code TEXT,
    estimated_delivery_date TIMESTAMPTZ,
    service_type TEXT,
    service_code TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TYPE order_status AS ENUM (
    'pending',
    'payment_success',
    'payment_failed',
    'shipped',
    'fulfillment_failed',
    'delivered',
    'cancelled',
    'return_requested',
    'returned',
    'refunded'
);

CREATE TABLE orders
(
    id UUID NOT NULL PRIMARY KEY,
    customer_id UUID NOT NULL REFERENCES customers (id),
    cart_id UUID NOT NULL REFERENCES carts (id),
    order_reference VARCHAR(255) UNIQUE NOT NULL,
    tax_amount NUMERIC(10, 2) NOT NULL,
    subtotal NUMERIC(10, 2) NOT NULL,
    total NUMERIC(10, 2) NOT NULL,
    currency VARCHAR(3) NOT NULL,
    tax_breakdown JSONB,
    shipping_rate NUMERIC(10, 2) NOT NULL,
    shipping_carrier_name TEXT,
    shipping_carrier_code TEXT,
    shipping_estimated_delivery_date TIMESTAMPTZ,
    shipping_service_type TEXT,
    shipping_service_code TEXT,
    delivery_full_name VARCHAR(255) NOT NULL,
    delivery_address VARCHAR(255) NOT NULL,
    delivery_apartment VARCHAR(255),
    delivery_city VARCHAR(255),
    delivery_state_code VARCHAR(255),
    delivery_country_code VARCHAR(255) NOT NULL,
    delivery_postal_code VARCHAR(255),
    delivery_phone_number VARCHAR(50),
    status ORDER_STATUS NOT NULL,
    fulfillment_message TEXT,
    fulfillment_shipment_date TIMESTAMPTZ,
    fulfillment_freight_charge NUMERIC(10, 2),
    fulfillment_order_total NUMERIC(10, 2),
    salesforce_id TEXT,
    stripe_payment_intent_id TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ
);

CREATE TABLE order_items
(
    id UUID NOT NULL PRIMARY KEY,
    order_id UUID NOT NULL REFERENCES orders (id),
    product_variant_id UUID NOT NULL REFERENCES product_variants (id),
    description TEXT,
    sku VARCHAR(255) NOT NULL,
    image_url TEXT,
    name VARCHAR(255) NOT NULL,
    length NUMERIC(10, 2),
    width NUMERIC(10, 2),
    height NUMERIC(10, 2),
    weight NUMERIC(10, 2),
    quantity INTEGER NOT NULL,
    price NUMERIC(10, 2) NOT NULL,
    salesforce_id TEXT,
    attributes JSONB,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ
);

-- +migrate Down
DROP TABLE IF EXISTS order_items;
DROP TABLE IF EXISTS orders;
DROP TABLE IF EXISTS cart_shipping_rates;
DROP TABLE IF EXISTS cart_items;
DROP TABLE IF EXISTS carts;
DROP TABLE IF EXISTS wishlist_items;
DROP TABLE IF EXISTS addresses;
DROP TABLE IF EXISTS customers;
DROP TABLE IF EXISTS product_variants;
DROP TABLE IF EXISTS products;
DROP TYPE IF EXISTS CART_STATUS;
DROP TYPE IF EXISTS ORDER_STATUS;
