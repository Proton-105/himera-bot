## Overview

Himera stores authoritative trading data in PostgreSQL. The database schema captures users, their token positions, and trade history. For the overall system architecture, refer to [docs/architecture.md](architecture.md).

## Schema

### users

Primary table for Telegram users interacting with the bot.

| Column      | Type           | Nullable | Default      | Notes                                |
|-------------|----------------|----------|--------------|--------------------------------------|
| telegram_id | BIGINT         | NO       | —            | Primary key; Telegram user identifier |
| username    | VARCHAR(255)   | YES      | —            | Telegram username (optional)         |
| balance     | DECIMAL(20,8)  | YES      | 10000        | User balance (non-negative)          |
| created_at  | TIMESTAMPTZ    | NO       | NOW()        | Creation timestamp (UTC)             |
| updated_at  | TIMESTAMPTZ    | NO       | NOW()        | Auto-updated via trigger             |

- Primary key: `telegram_id`.
- Trigger `users_set_updated_at` updates `updated_at` on every update.

### positions

Open token positions per user.

| Column       | Type           | Nullable | Default | Notes                                        |
|--------------|----------------|----------|---------|----------------------------------------------|
| id           | BIGSERIAL      | NO       | —       | Primary key                                  |
| telegram_id  | BIGINT         | NO       | —       | FK → `users(telegram_id)` (ON DELETE CASCADE) |
| token_address| VARCHAR(64)    | NO       | —       | Token contract address                       |
| token_symbol | VARCHAR(32)    | YES      | —       | Human-readable token symbol                  |
| amount       | DECIMAL(30,18) | NO       | —       | Position size (> 0)                          |
| avg_price    | DECIMAL(30,18) | NO       | —       | Average purchase price (> 0)                 |
| created_at   | TIMESTAMPTZ    | NO       | NOW()   | Creation timestamp (UTC)                     |

- Primary key: `id`.
- Indexes: `idx_positions_telegram_id` on `(telegram_id)` for per-user lookups.

### transactions

Historical trades executed by users.

| Column       | Type           | Nullable | Default | Notes                                         |
|--------------|----------------|----------|---------|-----------------------------------------------|
| id           | BIGSERIAL      | NO       | —       | Primary key                                   |
| telegram_id  | BIGINT         | NO       | —       | FK → `users(telegram_id)` (ON DELETE CASCADE)  |
| type         | VARCHAR(10)    | NO       | —       | `buy` or `sell` (CHECK constraint)            |
| token_address| VARCHAR(64)    | NO       | —       | Token contract address                        |
| amount       | DECIMAL(30,18) | NO       | —       | Trade amount (> 0)                            |
| price_usd    | DECIMAL(30,18) | NO       | —       | Unit price in USD (> 0)                       |
| total_usd    | DECIMAL(20,8)  | NO       | —       | Total trade value in USD                      |
| pnl_usd      | DECIMAL(20,8)  | YES      | —       | Profit/loss in USD                            |
| created_at   | TIMESTAMPTZ    | NO       | NOW()   | Timestamp of execution (UTC)                  |

- Primary key: `id`.
- Indexes:
  - `idx_transactions_telegram_id` on `(telegram_id)` for user history queries.
  - `idx_transactions_token_address` on `(token_address)` for asset-based analytics.

## Relationships

- `positions.telegram_id` → `users.telegram_id` (cascade delete). Removing a user cleans up positions automatically.
- `transactions.telegram_id` → `users.telegram_id` (cascade delete). Trade history is removed when the user is deleted.

These relationships ensure user-centric data integrity and simplify cleanup when accounts are removed.

## Conventions

- Table and column names use `snake_case`.
- Timestamps (`created_at`, `updated_at`) are stored in UTC (`TIMESTAMPTZ`).
- Monetary values are stored as fixed-precision decimals; floating-point types are not used to avoid rounding errors. Wherever possible, store amounts in minimal units (e.g., cents/wei) or enforce explicit precision via DECIMAL.

## Transactions

- All financial mutations (balances, positions, transactions) must execute inside explicit SQL transactions.
- Avoid invoking Telegram APIs, HTTP calls, or other external side effects within active transactions; commit first, then perform side effects (or use an outbox pattern).
- On failure within a transaction, the entire unit of work must roll back.
- See [docs/architecture.md](architecture.md) for the broader transaction strategy.

## Migration strategy

- Database migrations reside in the `migrations/` directory, versioned sequentially (`000001_...`, `000002_...`, etc.).
- Each migration should be idempotent where possible and provide a reversible counterpart (`down.sql`) when it is safe to do so.
- Schema changes go through Pull Requests and code review; never modify production databases manually.
