# Error Codes Reference

## Overview
Документация всех error codes, используемых в CHIMERA Sniper Bot.

## Error Categories

### E100 - Validation Errors
- **Code:** E100
- **Severity:** Low
- **Retryable:** No
- **User Message:** "Неверный формат данных"
- **Recovery:** Попросить пользователя повторить ввод
- **Examples:** Неверный формат суммы, некорректный адрес токена

### E200 - Database Errors
- **Code:** E200
- **Severity:** High
- **Retryable:** Yes (automatic)
- **User Message:** "Временная проблема, попробуйте позже"
- **Recovery:** Автоматический retry с exponential backoff (3 попытки)
- **Sentry:** Yes
- **Examples:** Deadlock, connection timeout, constraint violation

### E300 - External API Errors
- **Code:** E300
- **Severity:** Medium
- **Retryable:** Yes
- **User Message:** "Сервис временно недоступен"
- **Recovery:** Circuit breaker + retry
- **Examples:** DexScreener timeout, CoinGecko rate limit

### E400 - State Errors
- **Code:** E400
- **Severity:** Medium
- **Retryable:** No
- **User Message:** "Операция невозможна в текущем состоянии"
- **Recovery:** Clear state, return to Idle
- **Examples:** Invalid state transition, corrupted state data

### E500 - Rate Limit Errors
- **Code:** E500
- **Severity:** Low
- **Retryable:** No
- **User Message:** "Слишком много запросов. Попробуйте через X секунд"
- **Recovery:** Показать время ожидания
- **Examples:** User exceeded per-second limit

## Monitoring

Все ошибки с severity High/Critical автоматически отправляются в Sentry.
Для мониторинга используйте Prometheus метрики: `errors_total{code="E200"}`

## Testing

Для тестирования error handling используйте:
```go
err := errors.NewDatabaseError(sql.ErrConnDone)
userMsg, shouldRetry := handler.Handle(ctx, err)
```
