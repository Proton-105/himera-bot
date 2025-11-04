-- 000002_add_indexes.up.sql

CREATE INDEX IF NOT EXISTS idx_positions_telegram_id
    ON positions (telegram_id);

CREATE INDEX IF NOT EXISTS idx_transactions_telegram_id
    ON transactions (telegram_id);

CREATE INDEX IF NOT EXISTS idx_transactions_token_address
    ON transactions (token_address);
