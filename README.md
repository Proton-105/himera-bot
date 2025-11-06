# Himera Trading Bot

Himera — это бэкенд для трейдингового бота с Telegram-интерфейсом. Приложение агрегирует рыночные данные, управляет пользовательскими состояниями, хранит данные в PostgreSQL/Redis, публикует метрики Prometheus (`/metrics`) и health-check (`/health`) для мониторинга.

## Architecture overview

В проекте используется слоистая архитектура с чётким разделением ответственности.

```
                ┌────────────────┐
                │    cmd/bot     │  ← точка входа, сборка зависимостей
                └───────┬────────┘
                        │
              ┌─────────▼─────────┐
              │  internal/service  │  ← бизнес-логика, use-cases
              └─────────┬─────────┘
                        │
           ┌────────────▼────────────┐
           │    internal/domain      │  ← доменные сущности
           └────────────┬────────────┘
                        │
      ┌─────────────────▼─────────────────┐
      │        internal/repository        │  ← доступ к Postgres и Redis
      └───────────────┬───────────────────┘
                      │
        ┌─────────────▼─────────────┐
        │     internal/handler      │  ← Telegram/HTTP обработчики
        └─────────────┬─────────────┘
                      │
        ┌─────────────▼─────────────┐
        │      internal/state       │  ← FSM, фоновые задачи
        └─────────────┬─────────────┘
                      │
       ┌──────────────▼──────────────┐
       │ pkg/config │ pkg/logger │ pkg/redis │ … инфраструктура
       └─────────────────────────────┘
```

- `cmd/bot` — главный бинарь, собирает конфиг, инициализирует логгер, базы и Telegram-бота.
- `internal/domain` — доменные сущности (пользователь, состояния, трейды).
- `internal/service` — бизнес-логика и use cases.
- `internal/repository` — работа с PostgreSQL/Redis.
- `internal/handler` — обработчики транспортов (Telegram, HTTP).
- `internal/state` — FSM, фоновые задания, планировщики.
- `pkg/config`, `pkg/logger`, `pkg/redis` — инфраструктурные утилиты и адаптеры.

## Tech stack

- Go 1.24 (см. `go.mod`)
- PostgreSQL 15 (через `docker-compose`)
- Redis 7 (через `docker-compose`)
- Docker + docker-compose
- Telebot v3 (`gopkg.in/telebot.v3`)
- Prometheus `client_golang` для метрик
- Sentry SDK (`github.com/getsentry/sentry-go`)

## Prerequisites

- Go 1.24.x (`go env GOROOT` должен указывать на актуальную версию)
- Docker и docker-compose
- git
- Telegram Bot Token (передаётся через конфиг/переменные окружения)

## Setup & Run

1. **Клонируйте репозиторий**
   ```bash
   git clone https://github.com/Proton-105/himera-bot.git
   cd trading-bot
   ```

2. **Настройте конфигурацию**
   - Базовая конфигурация находится в `configs/config.yaml`.
   - Для переопределения значений используйте переменные окружения или дополнительные YAML-файлы (например, `configs/development.yaml`).
   - Ключевые параметры:
     - `server.port` — HTTP-порт основного сервера (обработчики, вебхуки).
     - `server.metrics_port` — порт для `/metrics` и `/health`.
     - `bot.token` — Telegram Bot Token (можно задать через YAML или переменную `BOT_TOKEN`).
     - `bot.mode` — `polling` или `webhook`.
     - `bot.webhook_url` — URL вебхука (для режима `webhook`).
     - `database.*` — параметры подключения к Postgres.
     - `redis.*` — параметры подключения к Redis.

   Таблица параметров:

   | Секция  | Ключ         | Описание                                | Значение по умолчанию |
   |---------|--------------|-----------------------------------------|-----------------------|
   | server  | port         | Основной HTTP-порт приложения           | `8080`                |
   | server  | metrics_port | Порт метрик и health-check              | `2112`                |
   | bot     | token        | Telegram Bot Token                      | `""`                  |
   | bot     | mode         | `polling` или `webhook`                 | `"polling"`           |
   | bot     | webhook_url  | URL для Telegram webhook                | `""`                  |
   | bot     | timeout      | Таймаут long polling                    | `120s`                |
   | database| host         | Хост PostgreSQL                         | `localhost`           |
   | database| port         | Порт PostgreSQL                         | `5434`                |
   | database| user         | Имя пользователя                        | `himera`              |
   | database| password     | Пароль                                  | `himera`              |
   | database| name         | Имя базы                                | `himera`              |
   | database| ssl_mode     | Режим SSL (`disable`, `require`, …)     | `disable`             |
   | redis   | host         | Хост Redis                              | `localhost`           |
   | redis   | port         | Порт Redis                              | `6380`                |
   | redis   | db           | Номер базы                              | `0`                   |
   | redis   | password     | Пароль                                  | `""`                  |

3. **Поднимите инфраструктуру**
   ```bash
   docker compose up -d
   ```
   Это создаст контейнеры PostgreSQL и Redis, определённые в `docker-compose.yml`.

4. **Запустите приложение**
   - Через Makefile:
     ```bash
     make run
     ```
   - Либо напрямую:
     ```bash
     go run ./cmd/bot
     ```

   При успешном старте:
   - Получится лог о запуске Telegram-бота.
   - `/metrics` и `/health` будут доступны на `http://localhost:<metrics_port>/`.

## Health & Metrics

- **/health** — HTTP endpoint, проверяет:
  - доступность PostgreSQL (`db.PingContext`);
  - ответ Redis (`PING`);
  - инициализацию Telegram Bot API.
  Возвращает JSON со статусами по компонентам. Если хотя бы одна проверка не `OK`, возвращается `503 Service Unavailable`.

- **/metrics** — Prometheus endpoint (`promhttp.Handler()`), отдаёт стандартные и кастомные метрики:
  - Redis-клиент (latency, ошибки);
  - HTTP-метрики (`promhttp`);
  - пользовательские метрики (по мере добавления функционала).

## Testing

| Команда                     | Назначение                                               |
|----------------------------|-----------------------------------------------------------|
| `go test ./...`            | Юнит-тесты приложения                                     |
| `golangci-lint run ./...`  | Статический анализ кода и линтинг                         |
| `pre-commit run --all-files` | Хуки форматирования, линтеры, дополнительные проверки |

Все команды можно запускать локально перед коммитом для быстрого фидбэка.

## Deployment

- CI/CD (GitHub Actions):
  - Линтинг (`golangci-lint`).
  - Тесты (`go test ./...`).
  - Покрытие (отчёты coverage).
  - Security-сканы (`gosec`/прочие проверки).
  - Сборка Docker-образа.
  - Деплой на staging/production (скрипты зависят от настроек окружения).

- Продакшн-деплой: сборка Docker-образа, обновление переменных окружения и конфигураций, раскатка в целевое окружение (Kubernetes/VM/Docker Swarm — в зависимости от выбранной инфраструктуры команды).

## Roadmap / Phases

- **Phase 0 — Bootstrap**: каркас приложения, инфраструктурные модули, health/metrics, базовые репозитории.
- **Phase 1 — MVP Trading Logic**: интеграция с DEX/аггрегаторами, первые торговые сценарии, расширение Telegram-команд.
- **Phase 2 — Advanced Automation**: стратегия, риск-менеджмент, оркестрация сделок, интеграция с внешними сигналами.
- **Phase 3 — Observability & Scaling**: расширенная телеметрия, алертинг, горизонтальное масштабирование, fault tolerance.

Готово! Теперь новый разработчик может поднять локальное окружение и познакомиться с кодовой базой за ~15 минут. Проверяйте README перед релизами, чтобы держать документацию актуальной.
