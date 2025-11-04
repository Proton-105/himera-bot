-- 000001_init_schema.down.sql

DROP TRIGGER IF EXISTS users_set_updated_at ON users;
DROP FUNCTION IF EXISTS set_updated_at();

DROP TABLE IF EXISTS transactions;
DROP TABLE IF EXISTS positions;
DROP TABLE IF EXISTS users;
