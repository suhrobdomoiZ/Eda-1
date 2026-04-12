CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- Пользователи
CREATE TABLE users
(
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    username      VARCHAR(128) NOT NULL CHECK (username != ''),
    password_hash VARCHAR(255) NOT NULL,
    role          VARCHAR(32)  NOT NULL CHECK (role IN ('user', 'admin', 'restaurant', 'courier'))
);

-- Профили ресторанов
CREATE TABLE restaurant_profiles
(
    user_id UUID PRIMARY KEY REFERENCES users (id) ON DELETE CASCADE,
    name    VARCHAR(256) NOT NULL CHECK (name != ''),
    address VARCHAR(512),
    phone   VARCHAR(32)
);

-- Профили курьеров
CREATE TABLE courier_profiles
(
    user_id UUID PRIMARY KEY REFERENCES users (id) ON DELETE CASCADE,
    name    VARCHAR(256) NOT NULL CHECK (name != ''),
    phone   VARCHAR(32)
);

-- Продукты ресторанов
CREATE TABLE products
(
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    restaurant_id UUID          REFERENCES users (id) ON DELETE CASCADE,
    name          VARCHAR(128)  NOT NULL CHECK (name != ''),
    description   VARCHAR(1024),
    price         NUMERIC(8, 2) NOT NULL CHECK (price > 0)
);

-- Заказы
CREATE TABLE orders
(
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    restaurant_id UUID         REFERENCES users (id) ON DELETE RESTRICT,
    courier_id    UUID         REFERENCES users (id) ON DELETE SET NULL,
    client_id     UUID         REFERENCES users (id) ON DELETE RESTRICT,
    address       VARCHAR(256) NOT NULL CHECK (address != ''),
    status        VARCHAR(64)  NOT NULL CHECK (status IN
                                ('created', 'cooking', 'ready', 'delivering', 'delivered', 'cancelled'))
);

-- Заказанные продукты
CREATE TABLE ordered_products
(
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    order_id   UUID    REFERENCES orders (id) ON DELETE CASCADE,
    product_id UUID    NOT NULL REFERENCES products (id) ON DELETE RESTRICT,
    count      INTEGER NOT NULL CHECK (count BETWEEN 1 AND 100)
);