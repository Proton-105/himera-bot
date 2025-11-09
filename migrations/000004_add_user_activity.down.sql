-- 000004_add_user_activity.down.sql

ALTER TABLE users
    DROP COLUMN IF EXISTS last_active_at,
    DROP COLUMN IF EXISTS is_blocked;
