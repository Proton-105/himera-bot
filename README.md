# Himera Trading Bot (Base L2) — Phase 0

Этот репозиторий — бэкенд Himera (trading bot для сети Base) в состоянии Phase 0:
каркас приложения, docker-окружение и минимальная схема БД под прототип.

## 1. Стек и структура

Язык: Go 1.22+
БД: PostgreSQL 15
Кэш: Redis 7
Контейнеры: Docker + docker compose

Основные директории:

- cmd/bot — входная точка приложения.
- internal/domain — доменные сущности (пока пусто).
- internal/service — use-cases (пока пусто).
- internal/repository — доступ к БД и кэшу (пока пусто).
- internal/handler — транспорты (Telegram, HTTP и т.п., пока пусто).
- internal/state — стейт-машина, фоновые джобы (пока пусто).
- internal/database — простой мигратор SQL-файлов.
- pkg/config — загрузка конфигурации из переменных окружения.
- pkg/logger — обёртка над log.Logger с префиксом [himera].
- migrations — SQL-миграции.
- scripts — служебные скрипты (init-db.sql, wait-for-it.sh, reset-db.sh).

## 2. Локальная разработка

Требования:

- Go 1.22+
- golangci-lint (через go install)
- pre-commit (через pip)

Базовые команды:

- make build — собрать бинарь bin/himera-bot.
- make run — запустить бота локально.
- make test — запустить go test ./...
- make lint — запустить golangci-lint run ./...
- pre-commit run --all-files — прогнать хуки локально.

## 3. Конфиг и переменные окружения

Шаблон окружения:

- файл .env.example в корне репозитория.

Создание локального файла:

- cp .env.example .env

Основные переменные:

- APP_ENV — окружение (development, staging, production).
- HTTP_PORT — порт HTTP-сервера бота.
- LOG_LEVEL — уровень логов.
- DB_HOST, DB_PORT, DB_USER, DB_PASSWORD, DB_NAME, DB_SSLMODE — настройки PostgreSQL.
- REDIS_ADDR, REDIS_DB — параметры Redis.

## 4. Docker и docker compose

Запуск стека (бот + Postgres + Redis):

1. Скопировать env:

- cp .env.example .env

2. Поднять контейнеры:

- docker compose up --build
- или docker compose up -d --build для запуска в фоне.

Сервисы:

- bot — контейнер с Himera trading bot.
- db — PostgreSQL 15 с volume db_data.
- redis — Redis 7.

Логи:

- docker compose logs -f bot
- docker compose logs -f db
- docker compose logs -f redis

Остановка:

- docker compose down
- docker compose down -v для полного сноса с volumes.

## 5. Схема БД (Phase 0)

Начальная схема:

Таблица users:

- telegram_id BIGINT PRIMARY KEY
- username VARCHAR(255)
- balance DECIMAL(20,8) DEFAULT 10000 CHECK (balance >= 0)
- created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
- updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()

Таблица positions:

- id BIGSERIAL PRIMARY KEY
- telegram_id BIGINT NOT NULL REFERENCES users(telegram_id) ON DELETE CASCADE
- token_address VARCHAR(64) NOT NULL
- token_symbol VARCHAR(32)
- amount DECIMAL(30,18) NOT NULL CHECK (amount > 0)
- avg_price DECIMAL(30,18) NOT NULL CHECK (avg_price > 0)
- UNIQUE (telegram_id, token_address)

Таблица transactions:

- id BIGSERIAL PRIMARY KEY
- telegram_id BIGINT NOT NULL REFERENCES users(telegram_id) ON DELETE CASCADE
- type VARCHAR(10) NOT NULL CHECK (type IN ('buy', 'sell'))
- token_address VARCHAR(64) NOT NULL
- amount DECIMAL(30,18) NOT NULL CHECK (amount > 0)
- price_usd DECIMAL(30,18) NOT NULL CHECK (price_usd > 0)
- total_usd DECIMAL(20,8) NOT NULL
- pnl_usd DECIMAL(20,8)
- created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()

## 6. Миграции

Файлы миграций:

- migrations/000001_init_schema.up.sql — создание таблиц и функции set_updated_at.
- migrations/000001_init_schema.down.sql — откат всей схемы users, positions, transactions.
- migrations/000002_add_indexes.up.sql — индексы:
  - idx_positions_telegram_id
  - idx_transactions_telegram_id
  - idx_transactions_token_address

Дополнительно есть scripts/init-db.sql с тем же initial-скриптом для быстрого наката.

## 7. Внутренний мигратор (internal/database/migrator.go)

Мигратор:

- читает директорию миграций;
- выбирает файлы с суффиксом .up.sql;
- сортирует их лексикографически;
- выполняет по очереди через db.ExecContext.

Пример использования (псевдокод, wiring будет в следующих фазах):

- открыть соединение к Postgres по DSN;
- создать logger;
- создать Migrator через NewMigrator;
- вызвать ApplyDir(ctx, "migrations").

## 8. Сброс БД (scripts/reset-db.sh)

Скрипт reset-db.sh:

- выполняет docker compose down -v;
- поднимает db и redis;
- ждёт готовности Postgres через wait-for-it.sh;
- поднимает bot.

Удобно для локальной разработки, чтобы быстро получить чистое окружение.

## 9. Дальнейшие фазы

Phase 1 и далее:

- описывают доменную модель,
- use-cases бота,
- Telegram-обработчики,
- интеграцию с сетью Base L2 и торговой логикой.

Phase 0 завершён, когда:

- lint и tests проходят локально,
- docker compose поднимает bot + db + redis,
- миграции успешно накатываются на чистую БД.
