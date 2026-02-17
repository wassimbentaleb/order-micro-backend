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

-- ─── Step 7: Bulk Seed Data ──────────────────────────────────

-- Additional roles
INSERT INTO user_schema.roles (name, description) VALUES
    ('moderator', 'Content moderator'),
    ('support', 'Customer support agent');

-- Additional categories
INSERT INTO product_schema.categories (name) VALUES
    ('Home & Garden'),
    ('Sports'),
    ('Toys'),
    ('Automotive'),
    ('Health'),
    ('Food'),
    ('Music'),
    ('Office');

-- ── Users (password_hash = bcrypt of "password123") ──
INSERT INTO user_schema.users (id, username, email, password_hash, created_at, updated_at) VALUES
    ('a0000001-0000-0000-0000-000000000001', 'alice',    'alice@example.com',    '$2b$10$4rl5yjF/boiHihMpktY5sOiexmy3WRmyLOgl9GUcvk6jUjZBDbaRy', NOW() - INTERVAL '90 days', NOW() - INTERVAL '1 day'),
    ('a0000001-0000-0000-0000-000000000002', 'bob',      'bob@example.com',      '$2b$10$4rl5yjF/boiHihMpktY5sOiexmy3WRmyLOgl9GUcvk6jUjZBDbaRy', NOW() - INTERVAL '85 days', NOW() - INTERVAL '2 days'),
    ('a0000001-0000-0000-0000-000000000003', 'charlie',  'charlie@example.com',  '$2b$10$4rl5yjF/boiHihMpktY5sOiexmy3WRmyLOgl9GUcvk6jUjZBDbaRy', NOW() - INTERVAL '80 days', NOW() - INTERVAL '3 days'),
    ('a0000001-0000-0000-0000-000000000004', 'diana',    'diana@example.com',    '$2b$10$4rl5yjF/boiHihMpktY5sOiexmy3WRmyLOgl9GUcvk6jUjZBDbaRy', NOW() - INTERVAL '75 days', NOW() - INTERVAL '5 days'),
    ('a0000001-0000-0000-0000-000000000005', 'edward',   'edward@example.com',   '$2b$10$4rl5yjF/boiHihMpktY5sOiexmy3WRmyLOgl9GUcvk6jUjZBDbaRy', NOW() - INTERVAL '70 days', NOW() - INTERVAL '1 day'),
    ('a0000001-0000-0000-0000-000000000006', 'fiona',    'fiona@example.com',    '$2b$10$4rl5yjF/boiHihMpktY5sOiexmy3WRmyLOgl9GUcvk6jUjZBDbaRy', NOW() - INTERVAL '65 days', NOW() - INTERVAL '10 days'),
    ('a0000001-0000-0000-0000-000000000007', 'george',   'george@example.com',   '$2b$10$4rl5yjF/boiHihMpktY5sOiexmy3WRmyLOgl9GUcvk6jUjZBDbaRy', NOW() - INTERVAL '60 days', NOW() - INTERVAL '4 days'),
    ('a0000001-0000-0000-0000-000000000008', 'hannah',   'hannah@example.com',   '$2b$10$4rl5yjF/boiHihMpktY5sOiexmy3WRmyLOgl9GUcvk6jUjZBDbaRy', NOW() - INTERVAL '55 days', NOW() - INTERVAL '2 days'),
    ('a0000001-0000-0000-0000-000000000009', 'ivan',     'ivan@example.com',     '$2b$10$4rl5yjF/boiHihMpktY5sOiexmy3WRmyLOgl9GUcvk6jUjZBDbaRy', NOW() - INTERVAL '50 days', NOW() - INTERVAL '7 days'),
    ('a0000001-0000-0000-0000-000000000010', 'julia',    'julia@example.com',    '$2b$10$4rl5yjF/boiHihMpktY5sOiexmy3WRmyLOgl9GUcvk6jUjZBDbaRy', NOW() - INTERVAL '45 days', NOW() - INTERVAL '1 day'),
    ('a0000001-0000-0000-0000-000000000011', 'kevin',    'kevin@example.com',    '$2b$10$4rl5yjF/boiHihMpktY5sOiexmy3WRmyLOgl9GUcvk6jUjZBDbaRy', NOW() - INTERVAL '40 days', NOW() - INTERVAL '3 days'),
    ('a0000001-0000-0000-0000-000000000012', 'laura',    'laura@example.com',    '$2b$10$4rl5yjF/boiHihMpktY5sOiexmy3WRmyLOgl9GUcvk6jUjZBDbaRy', NOW() - INTERVAL '35 days', NOW() - INTERVAL '6 days'),
    ('a0000001-0000-0000-0000-000000000013', 'mike',     'mike@example.com',     '$2b$10$4rl5yjF/boiHihMpktY5sOiexmy3WRmyLOgl9GUcvk6jUjZBDbaRy', NOW() - INTERVAL '30 days', NOW() - INTERVAL '2 days'),
    ('a0000001-0000-0000-0000-000000000014', 'natalie',  'natalie@example.com',  '$2b$10$4rl5yjF/boiHihMpktY5sOiexmy3WRmyLOgl9GUcvk6jUjZBDbaRy', NOW() - INTERVAL '25 days', NOW() - INTERVAL '1 day'),
    ('a0000001-0000-0000-0000-000000000015', 'oscar',    'oscar@example.com',    '$2b$10$4rl5yjF/boiHihMpktY5sOiexmy3WRmyLOgl9GUcvk6jUjZBDbaRy', NOW() - INTERVAL '20 days', NOW() - INTERVAL '5 days'),
    ('a0000001-0000-0000-0000-000000000016', 'patricia', 'patricia@example.com', '$2b$10$4rl5yjF/boiHihMpktY5sOiexmy3WRmyLOgl9GUcvk6jUjZBDbaRy', NOW() - INTERVAL '18 days', NOW() - INTERVAL '2 days'),
    ('a0000001-0000-0000-0000-000000000017', 'quinn',    'quinn@example.com',    '$2b$10$4rl5yjF/boiHihMpktY5sOiexmy3WRmyLOgl9GUcvk6jUjZBDbaRy', NOW() - INTERVAL '15 days', NOW() - INTERVAL '1 day'),
    ('a0000001-0000-0000-0000-000000000018', 'rachel',   'rachel@example.com',   '$2b$10$4rl5yjF/boiHihMpktY5sOiexmy3WRmyLOgl9GUcvk6jUjZBDbaRy', NOW() - INTERVAL '12 days', NOW() - INTERVAL '3 days'),
    ('a0000001-0000-0000-0000-000000000019', 'steve',    'steve@example.com',    '$2b$10$4rl5yjF/boiHihMpktY5sOiexmy3WRmyLOgl9GUcvk6jUjZBDbaRy', NOW() - INTERVAL '10 days', NOW() - INTERVAL '1 day'),
    ('a0000001-0000-0000-0000-000000000020', 'tina',     'tina@example.com',     '$2b$10$4rl5yjF/boiHihMpktY5sOiexmy3WRmyLOgl9GUcvk6jUjZBDbaRy', NOW() - INTERVAL '5 days',  NOW() - INTERVAL '1 day');

-- ── User Roles ──
INSERT INTO user_schema.user_roles (user_id, role_id) VALUES
    ('a0000001-0000-0000-0000-000000000001', 1), -- alice = admin
    ('a0000001-0000-0000-0000-000000000001', 2), -- alice = user
    ('a0000001-0000-0000-0000-000000000002', 2), -- bob = user
    ('a0000001-0000-0000-0000-000000000003', 2),
    ('a0000001-0000-0000-0000-000000000004', 2),
    ('a0000001-0000-0000-0000-000000000005', 3), -- edward = moderator
    ('a0000001-0000-0000-0000-000000000005', 2),
    ('a0000001-0000-0000-0000-000000000006', 2),
    ('a0000001-0000-0000-0000-000000000007', 4), -- george = support
    ('a0000001-0000-0000-0000-000000000007', 2),
    ('a0000001-0000-0000-0000-000000000008', 2),
    ('a0000001-0000-0000-0000-000000000009', 2),
    ('a0000001-0000-0000-0000-000000000010', 2),
    ('a0000001-0000-0000-0000-000000000011', 1), -- kevin = admin
    ('a0000001-0000-0000-0000-000000000011', 2),
    ('a0000001-0000-0000-0000-000000000012', 2),
    ('a0000001-0000-0000-0000-000000000013', 2),
    ('a0000001-0000-0000-0000-000000000014', 3), -- natalie = moderator
    ('a0000001-0000-0000-0000-000000000014', 2),
    ('a0000001-0000-0000-0000-000000000015', 2),
    ('a0000001-0000-0000-0000-000000000016', 2),
    ('a0000001-0000-0000-0000-000000000017', 2),
    ('a0000001-0000-0000-0000-000000000018', 4), -- rachel = support
    ('a0000001-0000-0000-0000-000000000018', 2),
    ('a0000001-0000-0000-0000-000000000019', 2),
    ('a0000001-0000-0000-0000-000000000020', 2);

-- ── Products (category_id: 1=Electronics, 2=Books, 3=Clothing, 4=Home&Garden, 5=Sports, 6=Toys, 7=Automotive, 8=Health, 9=Food, 10=Music, 11=Office) ──
INSERT INTO product_schema.products (id, name, description, price, category_id, created_at, updated_at) VALUES
    ('b0000001-0000-0000-0000-000000000001', 'Wireless Bluetooth Headphones',     'Noise-cancelling over-ear headphones with 30h battery',  89.99,  1, NOW() - INTERVAL '60 days', NOW() - INTERVAL '1 day'),
    ('b0000001-0000-0000-0000-000000000002', 'USB-C Charging Cable 2m',           'Braided nylon fast-charge cable',                        12.99,  1, NOW() - INTERVAL '58 days', NOW() - INTERVAL '2 days'),
    ('b0000001-0000-0000-0000-000000000003', 'Mechanical Keyboard RGB',           'Cherry MX Blue switches, full-size',                    149.99,  1, NOW() - INTERVAL '55 days', NOW() - INTERVAL '3 days'),
    ('b0000001-0000-0000-0000-000000000004', '27-inch 4K Monitor',                'IPS panel, 60Hz, USB-C input',                          349.99,  1, NOW() - INTERVAL '50 days', NOW() - INTERVAL '1 day'),
    ('b0000001-0000-0000-0000-000000000005', 'Wireless Mouse Ergonomic',          'Vertical design, 2.4GHz wireless',                       34.99,  1, NOW() - INTERVAL '48 days', NOW() - INTERVAL '5 days'),
    ('b0000001-0000-0000-0000-000000000006', 'Clean Code',                        'A handbook of agile software craftsmanship by Robert C. Martin', 39.99, 2, NOW() - INTERVAL '45 days', NOW() - INTERVAL '2 days'),
    ('b0000001-0000-0000-0000-000000000007', 'The Pragmatic Programmer',          'Journey to mastery, 20th anniversary edition',           44.99,  2, NOW() - INTERVAL '44 days', NOW() - INTERVAL '3 days'),
    ('b0000001-0000-0000-0000-000000000008', 'Design Patterns',                   'Elements of reusable object-oriented software',          49.99,  2, NOW() - INTERVAL '42 days', NOW() - INTERVAL '1 day'),
    ('b0000001-0000-0000-0000-000000000009', 'Introduction to Algorithms',        'Comprehensive algorithms textbook, 4th edition',         79.99,  2, NOW() - INTERVAL '40 days', NOW() - INTERVAL '4 days'),
    ('b0000001-0000-0000-0000-000000000010', 'Go Programming Language',           'Definitive guide to programming in Go',                  34.99,  2, NOW() - INTERVAL '38 days', NOW() - INTERVAL '2 days'),
    ('b0000001-0000-0000-0000-000000000011', 'Cotton T-Shirt Black',              '100% organic cotton, unisex fit',                        19.99,  3, NOW() - INTERVAL '35 days', NOW() - INTERVAL '1 day'),
    ('b0000001-0000-0000-0000-000000000012', 'Denim Jeans Slim Fit',              'Stretch denim, dark wash',                               59.99,  3, NOW() - INTERVAL '33 days', NOW() - INTERVAL '3 days'),
    ('b0000001-0000-0000-0000-000000000013', 'Winter Jacket Waterproof',          'Insulated, windproof, removable hood',                  129.99,  3, NOW() - INTERVAL '30 days', NOW() - INTERVAL '2 days'),
    ('b0000001-0000-0000-0000-000000000014', 'Running Sneakers',                  'Lightweight mesh, cushioned sole',                       79.99,  3, NOW() - INTERVAL '28 days', NOW() - INTERVAL '5 days'),
    ('b0000001-0000-0000-0000-000000000015', 'Wool Beanie Hat',                   'Merino wool, one size fits all',                         14.99,  3, NOW() - INTERVAL '25 days', NOW() - INTERVAL '1 day'),
    ('b0000001-0000-0000-0000-000000000016', 'LED Desk Lamp',                     'Adjustable brightness, USB charging port',               29.99,  4, NOW() - INTERVAL '22 days', NOW() - INTERVAL '2 days'),
    ('b0000001-0000-0000-0000-000000000017', 'Garden Tool Set 5-Piece',           'Stainless steel with wooden handles',                    45.99,  4, NOW() - INTERVAL '20 days', NOW() - INTERVAL '1 day'),
    ('b0000001-0000-0000-0000-000000000018', 'Yoga Mat Premium',                  'Non-slip, 6mm thick, carrying strap included',           24.99,  5, NOW() - INTERVAL '18 days', NOW() - INTERVAL '3 days'),
    ('b0000001-0000-0000-0000-000000000019', 'Dumbbell Set 20kg',                 'Adjustable weight, rubber coated',                       64.99,  5, NOW() - INTERVAL '15 days', NOW() - INTERVAL '2 days'),
    ('b0000001-0000-0000-0000-000000000020', 'Building Blocks 500pc',             'Compatible with major brands, assorted colors',           29.99,  6, NOW() - INTERVAL '12 days', NOW() - INTERVAL '1 day'),
    ('b0000001-0000-0000-0000-000000000021', 'Remote Control Car',                'Off-road 4WD, rechargeable battery',                     39.99,  6, NOW() - INTERVAL '10 days', NOW() - INTERVAL '2 days'),
    ('b0000001-0000-0000-0000-000000000022', 'Car Phone Mount',                   'Magnetic, 360-degree rotation',                          15.99,  7, NOW() - INTERVAL '8 days',  NOW() - INTERVAL '1 day'),
    ('b0000001-0000-0000-0000-000000000023', 'Vitamin D3 Supplements',            '1000 IU, 365 tablets',                                   12.99,  8, NOW() - INTERVAL '7 days',  NOW() - INTERVAL '1 day'),
    ('b0000001-0000-0000-0000-000000000024', 'Protein Powder Chocolate',          'Whey isolate, 2kg tub',                                  54.99,  8, NOW() - INTERVAL '6 days',  NOW() - INTERVAL '1 day'),
    ('b0000001-0000-0000-0000-000000000025', 'Organic Coffee Beans 1kg',          'Single origin, medium roast, whole bean',                22.99,  9, NOW() - INTERVAL '5 days',  NOW() - INTERVAL '1 day'),
    ('b0000001-0000-0000-0000-000000000026', 'Dark Chocolate Gift Box',           'Assorted truffles, 24 pieces',                           18.99,  9, NOW() - INTERVAL '4 days',  NOW() - INTERVAL '1 day'),
    ('b0000001-0000-0000-0000-000000000027', 'Acoustic Guitar Starter Pack',      'Full-size, includes tuner and picks',                   159.99, 10, NOW() - INTERVAL '3 days',  NOW() - INTERVAL '1 day'),
    ('b0000001-0000-0000-0000-000000000028', 'Standing Desk Converter',           'Height adjustable, fits on existing desk',               199.99, 11, NOW() - INTERVAL '2 days',  NOW() - INTERVAL '1 day'),
    ('b0000001-0000-0000-0000-000000000029', 'Ergonomic Office Chair',            'Lumbar support, mesh back, adjustable arms',             279.99, 11, NOW() - INTERVAL '1 day',   NOW()),
    ('b0000001-0000-0000-0000-000000000030', 'Portable Bluetooth Speaker',        'Waterproof IPX7, 12h battery life',                      49.99,  1, NOW() - INTERVAL '1 day',   NOW());

-- ── Inventory for all products ──
INSERT INTO product_schema.inventory (product_id, quantity, updated_at) VALUES
    ('b0000001-0000-0000-0000-000000000001', 150, NOW()),
    ('b0000001-0000-0000-0000-000000000002', 500, NOW()),
    ('b0000001-0000-0000-0000-000000000003', 75,  NOW()),
    ('b0000001-0000-0000-0000-000000000004', 40,  NOW()),
    ('b0000001-0000-0000-0000-000000000005', 200, NOW()),
    ('b0000001-0000-0000-0000-000000000006', 120, NOW()),
    ('b0000001-0000-0000-0000-000000000007', 90,  NOW()),
    ('b0000001-0000-0000-0000-000000000008', 60,  NOW()),
    ('b0000001-0000-0000-0000-000000000009', 45,  NOW()),
    ('b0000001-0000-0000-0000-000000000010', 110, NOW()),
    ('b0000001-0000-0000-0000-000000000011', 300, NOW()),
    ('b0000001-0000-0000-0000-000000000012', 180, NOW()),
    ('b0000001-0000-0000-0000-000000000013', 65,  NOW()),
    ('b0000001-0000-0000-0000-000000000014', 95,  NOW()),
    ('b0000001-0000-0000-0000-000000000015', 250, NOW()),
    ('b0000001-0000-0000-0000-000000000016', 130, NOW()),
    ('b0000001-0000-0000-0000-000000000017', 70,  NOW()),
    ('b0000001-0000-0000-0000-000000000018', 160, NOW()),
    ('b0000001-0000-0000-0000-000000000019', 55,  NOW()),
    ('b0000001-0000-0000-0000-000000000020', 200, NOW()),
    ('b0000001-0000-0000-0000-000000000021', 85,  NOW()),
    ('b0000001-0000-0000-0000-000000000022', 320, NOW()),
    ('b0000001-0000-0000-0000-000000000023', 400, NOW()),
    ('b0000001-0000-0000-0000-000000000024', 100, NOW()),
    ('b0000001-0000-0000-0000-000000000025', 140, NOW()),
    ('b0000001-0000-0000-0000-000000000026', 220, NOW()),
    ('b0000001-0000-0000-0000-000000000027', 30,  NOW()),
    ('b0000001-0000-0000-0000-000000000028', 50,  NOW()),
    ('b0000001-0000-0000-0000-000000000029', 35,  NOW()),
    ('b0000001-0000-0000-0000-000000000030', 175, NOW());

-- ── Orders ──
INSERT INTO order_schema.orders (id, user_id, status, total_amount, created_at, updated_at) VALUES
    ('c0000001-0000-0000-0000-000000000001', 'a0000001-0000-0000-0000-000000000001', 'completed',  102.98, NOW() - INTERVAL '80 days', NOW() - INTERVAL '78 days'),
    ('c0000001-0000-0000-0000-000000000002', 'a0000001-0000-0000-0000-000000000002', 'completed',  349.99, NOW() - INTERVAL '75 days', NOW() - INTERVAL '72 days'),
    ('c0000001-0000-0000-0000-000000000003', 'a0000001-0000-0000-0000-000000000003', 'completed',   84.98, NOW() - INTERVAL '70 days', NOW() - INTERVAL '68 days'),
    ('c0000001-0000-0000-0000-000000000004', 'a0000001-0000-0000-0000-000000000004', 'completed',  199.98, NOW() - INTERVAL '65 days', NOW() - INTERVAL '62 days'),
    ('c0000001-0000-0000-0000-000000000005', 'a0000001-0000-0000-0000-000000000005', 'completed',   59.98, NOW() - INTERVAL '60 days', NOW() - INTERVAL '58 days'),
    ('c0000001-0000-0000-0000-000000000006', 'a0000001-0000-0000-0000-000000000001', 'completed',  149.99, NOW() - INTERVAL '55 days', NOW() - INTERVAL '53 days'),
    ('c0000001-0000-0000-0000-000000000007', 'a0000001-0000-0000-0000-000000000006', 'completed',  129.99, NOW() - INTERVAL '50 days', NOW() - INTERVAL '48 days'),
    ('c0000001-0000-0000-0000-000000000008', 'a0000001-0000-0000-0000-000000000007', 'completed',   79.99, NOW() - INTERVAL '45 days', NOW() - INTERVAL '43 days'),
    ('c0000001-0000-0000-0000-000000000009', 'a0000001-0000-0000-0000-000000000008', 'completed',   64.99, NOW() - INTERVAL '40 days', NOW() - INTERVAL '38 days'),
    ('c0000001-0000-0000-0000-000000000010', 'a0000001-0000-0000-0000-000000000009', 'completed',   45.98, NOW() - INTERVAL '35 days', NOW() - INTERVAL '33 days'),
    ('c0000001-0000-0000-0000-000000000011', 'a0000001-0000-0000-0000-000000000010', 'completed',  279.99, NOW() - INTERVAL '30 days', NOW() - INTERVAL '28 days'),
    ('c0000001-0000-0000-0000-000000000012', 'a0000001-0000-0000-0000-000000000002', 'completed',   89.98, NOW() - INTERVAL '28 days', NOW() - INTERVAL '26 days'),
    ('c0000001-0000-0000-0000-000000000013', 'a0000001-0000-0000-0000-000000000011', 'completed',  159.99, NOW() - INTERVAL '25 days', NOW() - INTERVAL '23 days'),
    ('c0000001-0000-0000-0000-000000000014', 'a0000001-0000-0000-0000-000000000012', 'completed',   54.99, NOW() - INTERVAL '22 days', NOW() - INTERVAL '20 days'),
    ('c0000001-0000-0000-0000-000000000015', 'a0000001-0000-0000-0000-000000000013', 'shipped',    199.99, NOW() - INTERVAL '15 days', NOW() - INTERVAL '13 days'),
    ('c0000001-0000-0000-0000-000000000016', 'a0000001-0000-0000-0000-000000000014', 'shipped',     94.98, NOW() - INTERVAL '12 days', NOW() - INTERVAL '10 days'),
    ('c0000001-0000-0000-0000-000000000017', 'a0000001-0000-0000-0000-000000000015', 'shipped',    349.99, NOW() - INTERVAL '10 days', NOW() - INTERVAL '8 days'),
    ('c0000001-0000-0000-0000-000000000018', 'a0000001-0000-0000-0000-000000000003', 'processing', 109.98, NOW() - INTERVAL '7 days',  NOW() - INTERVAL '6 days'),
    ('c0000001-0000-0000-0000-000000000019', 'a0000001-0000-0000-0000-000000000016', 'processing',  49.99, NOW() - INTERVAL '5 days',  NOW() - INTERVAL '4 days'),
    ('c0000001-0000-0000-0000-000000000020', 'a0000001-0000-0000-0000-000000000017', 'pending',     22.99, NOW() - INTERVAL '3 days',  NOW() - INTERVAL '3 days'),
    ('c0000001-0000-0000-0000-000000000021', 'a0000001-0000-0000-0000-000000000018', 'pending',    159.99, NOW() - INTERVAL '2 days',  NOW() - INTERVAL '2 days'),
    ('c0000001-0000-0000-0000-000000000022', 'a0000001-0000-0000-0000-000000000019', 'pending',     34.99, NOW() - INTERVAL '1 day',   NOW() - INTERVAL '1 day'),
    ('c0000001-0000-0000-0000-000000000023', 'a0000001-0000-0000-0000-000000000020', 'pending',     79.99, NOW() - INTERVAL '1 day',   NOW()),
    ('c0000001-0000-0000-0000-000000000024', 'a0000001-0000-0000-0000-000000000001', 'cancelled',   29.99, NOW() - INTERVAL '20 days', NOW() - INTERVAL '19 days'),
    ('c0000001-0000-0000-0000-000000000025', 'a0000001-0000-0000-0000-000000000004', 'cancelled',   39.99, NOW() - INTERVAL '15 days', NOW() - INTERVAL '14 days');

-- ── Order Items ──
INSERT INTO order_schema.order_items (order_id, product_id, quantity, price) VALUES
    ('c0000001-0000-0000-0000-000000000001', 'b0000001-0000-0000-0000-000000000001', 1, 89.99),
    ('c0000001-0000-0000-0000-000000000001', 'b0000001-0000-0000-0000-000000000002', 1, 12.99),
    ('c0000001-0000-0000-0000-000000000002', 'b0000001-0000-0000-0000-000000000004', 1, 349.99),
    ('c0000001-0000-0000-0000-000000000003', 'b0000001-0000-0000-0000-000000000006', 1, 39.99),
    ('c0000001-0000-0000-0000-000000000003', 'b0000001-0000-0000-0000-000000000010', 1, 34.99),
    ('c0000001-0000-0000-0000-000000000003', 'b0000001-0000-0000-0000-000000000015', 1, 14.99),
    ('c0000001-0000-0000-0000-000000000004', 'b0000001-0000-0000-0000-000000000011', 2, 19.99),
    ('c0000001-0000-0000-0000-000000000004', 'b0000001-0000-0000-0000-000000000013', 1, 129.99),
    ('c0000001-0000-0000-0000-000000000004', 'b0000001-0000-0000-0000-000000000015', 2, 14.99),
    ('c0000001-0000-0000-0000-000000000005', 'b0000001-0000-0000-0000-000000000012', 1, 59.99),
    ('c0000001-0000-0000-0000-000000000006', 'b0000001-0000-0000-0000-000000000003', 1, 149.99),
    ('c0000001-0000-0000-0000-000000000007', 'b0000001-0000-0000-0000-000000000013', 1, 129.99),
    ('c0000001-0000-0000-0000-000000000008', 'b0000001-0000-0000-0000-000000000014', 1, 79.99),
    ('c0000001-0000-0000-0000-000000000009', 'b0000001-0000-0000-0000-000000000019', 1, 64.99),
    ('c0000001-0000-0000-0000-000000000010', 'b0000001-0000-0000-0000-000000000025', 2, 22.99),
    ('c0000001-0000-0000-0000-000000000011', 'b0000001-0000-0000-0000-000000000029', 1, 279.99),
    ('c0000001-0000-0000-0000-000000000012', 'b0000001-0000-0000-0000-000000000001', 1, 89.99),
    ('c0000001-0000-0000-0000-000000000013', 'b0000001-0000-0000-0000-000000000027', 1, 159.99),
    ('c0000001-0000-0000-0000-000000000014', 'b0000001-0000-0000-0000-000000000024', 1, 54.99),
    ('c0000001-0000-0000-0000-000000000015', 'b0000001-0000-0000-0000-000000000028', 1, 199.99),
    ('c0000001-0000-0000-0000-000000000016', 'b0000001-0000-0000-0000-000000000005', 1, 34.99),
    ('c0000001-0000-0000-0000-000000000016', 'b0000001-0000-0000-0000-000000000012', 1, 59.99),
    ('c0000001-0000-0000-0000-000000000017', 'b0000001-0000-0000-0000-000000000004', 1, 349.99),
    ('c0000001-0000-0000-0000-000000000018', 'b0000001-0000-0000-0000-000000000001', 1, 89.99),
    ('c0000001-0000-0000-0000-000000000018', 'b0000001-0000-0000-0000-000000000011', 1, 19.99),
    ('c0000001-0000-0000-0000-000000000019', 'b0000001-0000-0000-0000-000000000030', 1, 49.99),
    ('c0000001-0000-0000-0000-000000000020', 'b0000001-0000-0000-0000-000000000025', 1, 22.99),
    ('c0000001-0000-0000-0000-000000000021', 'b0000001-0000-0000-0000-000000000027', 1, 159.99),
    ('c0000001-0000-0000-0000-000000000022', 'b0000001-0000-0000-0000-000000000005', 1, 34.99),
    ('c0000001-0000-0000-0000-000000000023', 'b0000001-0000-0000-0000-000000000014', 1, 79.99),
    ('c0000001-0000-0000-0000-000000000024', 'b0000001-0000-0000-0000-000000000016', 1, 29.99),
    ('c0000001-0000-0000-0000-000000000025', 'b0000001-0000-0000-0000-000000000021', 1, 39.99);

-- ── Payments ──
INSERT INTO order_schema.payments (id, order_id, method, status, paid_at) VALUES
    ('d0000001-0000-0000-0000-000000000001', 'c0000001-0000-0000-0000-000000000001', 'credit_card', 'completed', NOW() - INTERVAL '80 days'),
    ('d0000001-0000-0000-0000-000000000002', 'c0000001-0000-0000-0000-000000000002', 'paypal',      'completed', NOW() - INTERVAL '75 days'),
    ('d0000001-0000-0000-0000-000000000003', 'c0000001-0000-0000-0000-000000000003', 'credit_card', 'completed', NOW() - INTERVAL '70 days'),
    ('d0000001-0000-0000-0000-000000000004', 'c0000001-0000-0000-0000-000000000004', 'debit_card',  'completed', NOW() - INTERVAL '65 days'),
    ('d0000001-0000-0000-0000-000000000005', 'c0000001-0000-0000-0000-000000000005', 'credit_card', 'completed', NOW() - INTERVAL '60 days'),
    ('d0000001-0000-0000-0000-000000000006', 'c0000001-0000-0000-0000-000000000006', 'paypal',      'completed', NOW() - INTERVAL '55 days'),
    ('d0000001-0000-0000-0000-000000000007', 'c0000001-0000-0000-0000-000000000007', 'credit_card', 'completed', NOW() - INTERVAL '50 days'),
    ('d0000001-0000-0000-0000-000000000008', 'c0000001-0000-0000-0000-000000000008', 'debit_card',  'completed', NOW() - INTERVAL '45 days'),
    ('d0000001-0000-0000-0000-000000000009', 'c0000001-0000-0000-0000-000000000009', 'credit_card', 'completed', NOW() - INTERVAL '40 days'),
    ('d0000001-0000-0000-0000-000000000010', 'c0000001-0000-0000-0000-000000000010', 'paypal',      'completed', NOW() - INTERVAL '35 days'),
    ('d0000001-0000-0000-0000-000000000011', 'c0000001-0000-0000-0000-000000000011', 'credit_card', 'completed', NOW() - INTERVAL '30 days'),
    ('d0000001-0000-0000-0000-000000000012', 'c0000001-0000-0000-0000-000000000012', 'debit_card',  'completed', NOW() - INTERVAL '28 days'),
    ('d0000001-0000-0000-0000-000000000013', 'c0000001-0000-0000-0000-000000000013', 'paypal',      'completed', NOW() - INTERVAL '25 days'),
    ('d0000001-0000-0000-0000-000000000014', 'c0000001-0000-0000-0000-000000000014', 'credit_card', 'completed', NOW() - INTERVAL '22 days'),
    ('d0000001-0000-0000-0000-000000000015', 'c0000001-0000-0000-0000-000000000015', 'credit_card', 'completed', NOW() - INTERVAL '15 days'),
    ('d0000001-0000-0000-0000-000000000016', 'c0000001-0000-0000-0000-000000000016', 'paypal',      'completed', NOW() - INTERVAL '12 days'),
    ('d0000001-0000-0000-0000-000000000017', 'c0000001-0000-0000-0000-000000000017', 'debit_card',  'completed', NOW() - INTERVAL '10 days'),
    ('d0000001-0000-0000-0000-000000000018', 'c0000001-0000-0000-0000-000000000018', 'credit_card', 'completed', NOW() - INTERVAL '7 days'),
    ('d0000001-0000-0000-0000-000000000019', 'c0000001-0000-0000-0000-000000000019', 'paypal',      'pending',   NULL),
    ('d0000001-0000-0000-0000-000000000020', 'c0000001-0000-0000-0000-000000000020', 'credit_card', 'pending',   NULL),
    ('d0000001-0000-0000-0000-000000000021', 'c0000001-0000-0000-0000-000000000021', 'debit_card',  'pending',   NULL),
    ('d0000001-0000-0000-0000-000000000022', 'c0000001-0000-0000-0000-000000000022', 'credit_card', 'pending',   NULL),
    ('d0000001-0000-0000-0000-000000000023', 'c0000001-0000-0000-0000-000000000023', 'paypal',      'pending',   NULL),
    ('d0000001-0000-0000-0000-000000000024', 'c0000001-0000-0000-0000-000000000024', 'credit_card', 'refunded',  NOW() - INTERVAL '19 days'),
    ('d0000001-0000-0000-0000-000000000025', 'c0000001-0000-0000-0000-000000000025', 'debit_card',  'refunded',  NOW() - INTERVAL '14 days');

-- ── Notification Logs ──
INSERT INTO notification_schema.notif_logs (user_id, type, subject, body, status, sent_at) VALUES
    ('a0000001-0000-0000-0000-000000000001', 'email', 'Welcome!',                    'Welcome to our platform, alice!',                         'sent',   NOW() - INTERVAL '90 days'),
    ('a0000001-0000-0000-0000-000000000002', 'email', 'Welcome!',                    'Welcome to our platform, bob!',                           'sent',   NOW() - INTERVAL '85 days'),
    ('a0000001-0000-0000-0000-000000000003', 'email', 'Welcome!',                    'Welcome to our platform, charlie!',                       'sent',   NOW() - INTERVAL '80 days'),
    ('a0000001-0000-0000-0000-000000000004', 'email', 'Welcome!',                    'Welcome to our platform, diana!',                         'sent',   NOW() - INTERVAL '75 days'),
    ('a0000001-0000-0000-0000-000000000005', 'email', 'Welcome!',                    'Welcome to our platform, edward!',                        'sent',   NOW() - INTERVAL '70 days'),
    ('a0000001-0000-0000-0000-000000000006', 'email', 'Welcome!',                    'Welcome to our platform, fiona!',                         'sent',   NOW() - INTERVAL '65 days'),
    ('a0000001-0000-0000-0000-000000000007', 'email', 'Welcome!',                    'Welcome to our platform, george!',                        'sent',   NOW() - INTERVAL '60 days'),
    ('a0000001-0000-0000-0000-000000000008', 'email', 'Welcome!',                    'Welcome to our platform, hannah!',                        'sent',   NOW() - INTERVAL '55 days'),
    ('a0000001-0000-0000-0000-000000000009', 'email', 'Welcome!',                    'Welcome to our platform, ivan!',                          'sent',   NOW() - INTERVAL '50 days'),
    ('a0000001-0000-0000-0000-000000000010', 'email', 'Welcome!',                    'Welcome to our platform, julia!',                         'sent',   NOW() - INTERVAL '45 days'),
    ('a0000001-0000-0000-0000-000000000001', 'email', 'Order Confirmed',             'Your order #c0000001-...-01 has been placed successfully.','sent',   NOW() - INTERVAL '80 days'),
    ('a0000001-0000-0000-0000-000000000002', 'email', 'Order Confirmed',             'Your order #c0000001-...-02 has been placed successfully.','sent',   NOW() - INTERVAL '75 days'),
    ('a0000001-0000-0000-0000-000000000003', 'email', 'Order Confirmed',             'Your order #c0000001-...-03 has been placed successfully.','sent',   NOW() - INTERVAL '70 days'),
    ('a0000001-0000-0000-0000-000000000004', 'email', 'Order Confirmed',             'Your order #c0000001-...-04 has been placed successfully.','sent',   NOW() - INTERVAL '65 days'),
    ('a0000001-0000-0000-0000-000000000005', 'email', 'Order Confirmed',             'Your order #c0000001-...-05 has been placed successfully.','sent',   NOW() - INTERVAL '60 days'),
    ('a0000001-0000-0000-0000-000000000001', 'email', 'Order Delivered',             'Your order #c0000001-...-01 has been delivered.',          'sent',   NOW() - INTERVAL '78 days'),
    ('a0000001-0000-0000-0000-000000000002', 'email', 'Order Delivered',             'Your order #c0000001-...-02 has been delivered.',          'sent',   NOW() - INTERVAL '72 days'),
    ('a0000001-0000-0000-0000-000000000003', 'email', 'Order Delivered',             'Your order #c0000001-...-03 has been delivered.',          'sent',   NOW() - INTERVAL '68 days'),
    ('a0000001-0000-0000-0000-000000000006', 'email', 'Order Confirmed',             'Your order #c0000001-...-07 has been placed successfully.','sent',   NOW() - INTERVAL '50 days'),
    ('a0000001-0000-0000-0000-000000000007', 'email', 'Order Confirmed',             'Your order #c0000001-...-08 has been placed successfully.','sent',   NOW() - INTERVAL '45 days'),
    ('a0000001-0000-0000-0000-000000000008', 'email', 'Order Confirmed',             'Your order #c0000001-...-09 has been placed successfully.','sent',   NOW() - INTERVAL '40 days'),
    ('a0000001-0000-0000-0000-000000000009', 'email', 'Order Confirmed',             'Your order #c0000001-...-10 has been placed successfully.','sent',   NOW() - INTERVAL '35 days'),
    ('a0000001-0000-0000-0000-000000000010', 'email', 'Order Confirmed',             'Your order #c0000001-...-11 has been placed successfully.','sent',   NOW() - INTERVAL '30 days'),
    ('a0000001-0000-0000-0000-000000000011', 'email', 'Welcome!',                    'Welcome to our platform, kevin!',                         'sent',   NOW() - INTERVAL '40 days'),
    ('a0000001-0000-0000-0000-000000000012', 'email', 'Welcome!',                    'Welcome to our platform, laura!',                         'sent',   NOW() - INTERVAL '35 days'),
    ('a0000001-0000-0000-0000-000000000013', 'email', 'Welcome!',                    'Welcome to our platform, mike!',                          'sent',   NOW() - INTERVAL '30 days'),
    ('a0000001-0000-0000-0000-000000000014', 'email', 'Welcome!',                    'Welcome to our platform, natalie!',                       'sent',   NOW() - INTERVAL '25 days'),
    ('a0000001-0000-0000-0000-000000000015', 'email', 'Welcome!',                    'Welcome to our platform, oscar!',                         'sent',   NOW() - INTERVAL '20 days'),
    ('a0000001-0000-0000-0000-000000000016', 'email', 'Welcome!',                    'Welcome to our platform, patricia!',                      'sent',   NOW() - INTERVAL '18 days'),
    ('a0000001-0000-0000-0000-000000000017', 'email', 'Welcome!',                    'Welcome to our platform, quinn!',                         'sent',   NOW() - INTERVAL '15 days'),
    ('a0000001-0000-0000-0000-000000000018', 'email', 'Welcome!',                    'Welcome to our platform, rachel!',                        'sent',   NOW() - INTERVAL '12 days'),
    ('a0000001-0000-0000-0000-000000000019', 'email', 'Welcome!',                    'Welcome to our platform, steve!',                         'sent',   NOW() - INTERVAL '10 days'),
    ('a0000001-0000-0000-0000-000000000020', 'email', 'Welcome!',                    'Welcome to our platform, tina!',                          'sent',   NOW() - INTERVAL '5 days'),
    ('a0000001-0000-0000-0000-000000000001', 'email', 'Stock Alert',                 'Product Acoustic Guitar Starter Pack is running low.',     'sent',   NOW() - INTERVAL '3 days'),
    ('a0000001-0000-0000-0000-000000000011', 'email', 'Order Confirmed',             'Your order #c0000001-...-13 has been placed successfully.','sent',   NOW() - INTERVAL '25 days'),
    ('a0000001-0000-0000-0000-000000000012', 'email', 'Order Confirmed',             'Your order #c0000001-...-14 has been placed successfully.','sent',   NOW() - INTERVAL '22 days'),
    ('a0000001-0000-0000-0000-000000000017', 'email', 'Order Confirmed',             'Your order #c0000001-...-20 has been placed successfully.','sent',   NOW() - INTERVAL '3 days'),
    ('a0000001-0000-0000-0000-000000000018', 'email', 'Order Confirmed',             'Your order #c0000001-...-21 has been placed successfully.','sent',   NOW() - INTERVAL '2 days'),
    ('a0000001-0000-0000-0000-000000000019', 'email', 'Order Confirmed',             'Your order #c0000001-...-22 has been placed successfully.','sent',   NOW() - INTERVAL '1 day'),
    ('a0000001-0000-0000-0000-000000000020', 'email', 'Order Confirmed',             'Your order #c0000001-...-23 has been placed successfully.','sent',   NOW());
