-- ============================================================
-- Microservice Database Initialization
-- Runs once on first PostgreSQL startup
-- ============================================================

-- ─── Step 1: Create Schemas ─────────────────────────────────
CREATE SCHEMA IF NOT EXISTS user_schema;
CREATE SCHEMA IF NOT EXISTS product_schema;
CREATE SCHEMA IF NOT EXISTS order_schema;
CREATE SCHEMA IF NOT EXISTS notification_schema;

-- ─── Step 2: Create Service DB Users ────────────────────────
DO $$
BEGIN
    IF NOT EXISTS (SELECT FROM pg_catalog.pg_roles WHERE rolname = 'svc_user') THEN
        CREATE USER svc_user WITH PASSWORD 'svc_user_pass';
    END IF;
    IF NOT EXISTS (SELECT FROM pg_catalog.pg_roles WHERE rolname = 'svc_product') THEN
        CREATE USER svc_product WITH PASSWORD 'svc_product_pass';
    END IF;
    IF NOT EXISTS (SELECT FROM pg_catalog.pg_roles WHERE rolname = 'svc_order') THEN
        CREATE USER svc_order WITH PASSWORD 'svc_order_pass';
    END IF;
    IF NOT EXISTS (SELECT FROM pg_catalog.pg_roles WHERE rolname = 'svc_notif') THEN
        CREATE USER svc_notif WITH PASSWORD 'svc_notif_pass';
    END IF;
END
$$;

-- ─── Step 3: Grant Permissions ──────────────────────────────

-- svc_user → user_schema only
GRANT USAGE ON SCHEMA user_schema TO svc_user;
GRANT CREATE ON SCHEMA user_schema TO svc_user;
ALTER DEFAULT PRIVILEGES IN SCHEMA user_schema GRANT ALL PRIVILEGES ON TABLES TO svc_user;
ALTER DEFAULT PRIVILEGES IN SCHEMA user_schema GRANT ALL PRIVILEGES ON SEQUENCES TO svc_user;

-- svc_product → product_schema only
GRANT USAGE ON SCHEMA product_schema TO svc_product;
GRANT CREATE ON SCHEMA product_schema TO svc_product;
ALTER DEFAULT PRIVILEGES IN SCHEMA product_schema GRANT ALL PRIVILEGES ON TABLES TO svc_product;
ALTER DEFAULT PRIVILEGES IN SCHEMA product_schema GRANT ALL PRIVILEGES ON SEQUENCES TO svc_product;

-- svc_order → order_schema only
GRANT USAGE ON SCHEMA order_schema TO svc_order;
GRANT CREATE ON SCHEMA order_schema TO svc_order;
ALTER DEFAULT PRIVILEGES IN SCHEMA order_schema GRANT ALL PRIVILEGES ON TABLES TO svc_order;
ALTER DEFAULT PRIVILEGES IN SCHEMA order_schema GRANT ALL PRIVILEGES ON SEQUENCES TO svc_order;

-- svc_notif → notification_schema only
GRANT USAGE ON SCHEMA notification_schema TO svc_notif;
GRANT CREATE ON SCHEMA notification_schema TO svc_notif;
ALTER DEFAULT PRIVILEGES IN SCHEMA notification_schema GRANT ALL PRIVILEGES ON TABLES TO svc_notif;
ALTER DEFAULT PRIVILEGES IN SCHEMA notification_schema GRANT ALL PRIVILEGES ON SEQUENCES TO svc_notif;

-- ─── Step 4: Create Tables ──────────────────────────────────

-- ==================== user_schema ====================

CREATE TABLE user_schema.roles (
    id SERIAL PRIMARY KEY,
    name VARCHAR(50) UNIQUE NOT NULL,
    description TEXT
);

CREATE TABLE user_schema.users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    username VARCHAR(100) UNIQUE NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

CREATE TABLE user_schema.user_roles (
    user_id UUID REFERENCES user_schema.users(id) ON DELETE CASCADE,
    role_id INT REFERENCES user_schema.roles(id) ON DELETE CASCADE,
    PRIMARY KEY (user_id, role_id)
);

-- ==================== product_schema ====================

CREATE TABLE product_schema.categories (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) UNIQUE NOT NULL
);

CREATE TABLE product_schema.products (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(200) NOT NULL,
    description TEXT,
    price DECIMAL(10, 2) NOT NULL,
    category_id INT REFERENCES product_schema.categories(id) ON DELETE SET NULL,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

CREATE TABLE product_schema.inventory (
    id SERIAL PRIMARY KEY,
    product_id UUID UNIQUE REFERENCES product_schema.products(id) ON DELETE CASCADE,
    quantity INT NOT NULL DEFAULT 0,
    updated_at TIMESTAMP DEFAULT NOW()
);

-- ==================== order_schema ====================

CREATE TABLE order_schema.orders (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'pending',
    total_amount DECIMAL(10, 2) NOT NULL,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

CREATE TABLE order_schema.order_items (
    id SERIAL PRIMARY KEY,
    order_id UUID REFERENCES order_schema.orders(id) ON DELETE CASCADE,
    product_id UUID NOT NULL,
    quantity INT NOT NULL,
    price DECIMAL(10, 2) NOT NULL
);

CREATE TABLE order_schema.payments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    order_id UUID REFERENCES order_schema.orders(id) ON DELETE CASCADE,
    method VARCHAR(50) NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'pending',
    paid_at TIMESTAMP
);

-- ==================== notification_schema ====================

CREATE TABLE notification_schema.templates (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) UNIQUE NOT NULL,
    body_template TEXT NOT NULL,
    type VARCHAR(20) NOT NULL
);

CREATE TABLE notification_schema.notif_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL,
    type VARCHAR(20) NOT NULL,
    subject VARCHAR(255),
    body TEXT,
    status VARCHAR(20) DEFAULT 'sent',
    sent_at TIMESTAMP DEFAULT NOW()
);

-- ─── Step 5: Grant Privileges on Created Tables ─────────────
-- (needed because tables were created by postgres user, not service users)

GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA user_schema TO svc_user;
GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA user_schema TO svc_user;

GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA product_schema TO svc_product;
GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA product_schema TO svc_product;

GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA order_schema TO svc_order;
GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA order_schema TO svc_order;

GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA notification_schema TO svc_notif;
GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA notification_schema TO svc_notif;

-- ─── Step 6: Seed Data ──────────────────────────────────────

-- Default roles
INSERT INTO user_schema.roles (name, description) VALUES
    ('admin', 'Administrator with full access'),
    ('user', 'Regular user');

-- Default categories
INSERT INTO product_schema.categories (name) VALUES
    ('Electronics'),
    ('Books'),
    ('Clothing');

-- Default notification templates
INSERT INTO notification_schema.templates (name, body_template, type) VALUES
    ('welcome_email', 'Welcome to our platform, {{username}}!', 'email'),
    ('order_confirmation', 'Your order #{{order_id}} has been placed successfully.', 'email'),
    ('order_completed', 'Your order #{{order_id}} has been delivered.', 'email'),
    ('stock_alert', 'Product {{product_name}} is out of stock.', 'email');
