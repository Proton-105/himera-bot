-- 000001_init_schema.up.sql

CREATE TABLE IF NOT EXISTS users (
    telegram_id BIGINT PRIMARY KEY,
    username VARCHAR(255),
    balance DECIMAL(20,8) DEFAULT 10000 CHECK (balance >= 0),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS positions (
    id BIGSERIAL PRIMARY KEY,
    telegram_id BIGINT NOT NULL REFERENCES users(telegram_id) ON DELETE CASCADE,
    token_address VARCHAR(64) NOT NULL,
    token_symbol VARCHAR(32),
    amount DECIMAL(30,18) NOT NULL CHECK (amount > 0),
    avg_price DECIMAL(30,18) NOT NULL CHECK (avg_price > 0),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS transactions (
    id BIGSERIAL PRIMARY KEY,
    telegram_id BIGINT NOT NULL REFERENCES users(telegram_id) ON DELETE CASCADE,
    type VARCHAR(10) NOT NULL CHECK (type IN ('buy', 'sell')),
    token_address VARCHAR(64) NOT NULL,
    amount DECIMAL(30,18) NOT NULL CHECK (amount > 0),
    price_usd DECIMAL(30,18) NOT NULL CHECK (price_usd > 0),
    total_usd DECIMAL(20,8) NOT NULL,
    pnl_usd DECIMAL(20,8),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE OR REPLACE FUNCTION set_updated_at()
RETURNS trigger AS $$
BEGIN
    NEW.updated_at := NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER users_set_updated_at
BEFORE UPDATE ON users
FOR EACH ROW
EXECUTE FUNCTION set_updated_at();
