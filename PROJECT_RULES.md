# Himera Trading Bot — Project Rules for LLM Assistants (with MCP)

This document defines how LLM assistants (Cursor, Claude, ChatGPT, etc.) must operate in the **Himera** trading bot project.

---

## 1. General principles

- Language: **Go**.
- Architecture: **Clean / Hexagonal**:
  - `cmd/bot` – entrypoints.
  - `internal/domain` – entities & core logic.
  - `internal/service` – use cases / application services.
  - `internal/repository` – DB/cache access (Postgres, Redis, etc.).
  - `internal/handler` – transports (Telegram Bot API, HTTP/GRPC, workers).
  - `internal/state` – state machine, background jobs, schedulers.
  - `pkg/*` – shared utils (logger, config, clients, etc.).
- Code, identifiers, code comments — **English**. Explanations to the author may be in Russian.

---

## 2. Editing rules

- **Always rewrite the entire file.**
  No diffs, no partial edits, no `...`. Output full, consistent file content each time.
- Do **not** change the tech stack (Go, Postgres, Redis, etc.) without explicit request.
- Code must be formatted with `gofmt`.
- Use idiomatic Go error handling:
  - return `error` values;
  - no `panic` in business logic (only in `main`/bootstrap if absolutely necessary).
- When changing structs/interfaces/types:
  - update all usages;
  - keep the project compilable.
- Money-related logic:
  - never use `float64` for balances, prices, P&L;
  - use integers (minimal units) or decimal/big-int;
  - make units explicit in names/comments (`base_wei`, `usd_cents`, etc.).

---

## 3. Phases & micro-steps

Work is organized by phases and micro-steps, e.g. `Phase2-M003`.

For each micro-step, the assistant must:

1. **Read** the user’s task description for that step.
2. **Propose a short plan** (2–7 bullets):
   - which files will be created/overwritten;
   - which structs/interfaces will be added or modified;
   - which external systems are touched (DB, Redis, Telegram, RPC).
3. **Implement only the scope of this step.**
   No additional features or refactors unless explicitly requested.
4. **Provide verification commands**, e.g.:
   - `go build ./cmd/bot`
   - `go test ./...`
   - `docker compose up -d`
5. Optionally list edge cases and future improvements, **without** implementing them.

If the step targets a specific module, do not modify unrelated modules unless absolutely required.

---

## 4. Configuration & secrets

- All sensitive data comes from **environment variables** / `.env` files (not committed).
- Do not hard-code:
  - API keys,
  - private keys,
  - DB credentials, tokens, etc.
- `pkg/config` should:
  - define a clear `Config` struct (with nested sections where needed);
  - validate critical fields and fail fast if they are missing.

---

## 5. Logging & observability

- Logs must be structured and meaningful:
  - include context (user_id, chat_id, account_id, tx_hash, etc.) where relevant;
  - use clear levels (info, warn, error).
- Errors:
  - must be logged with context;
  - should be propagated up the stack when appropriate;
  - must not be silently ignored.
- Future integration with:
  - **Prometheus** for metrics,
  - **Sentry** (or similar) for error tracking,
  should be kept in mind when adding logging/metrics hooks.

---

## 6. Database & transactions

- Any operation that affects money must be:
  - **atomic**;
  - preferably wrapped in a DB transaction.
- Avoid external side-effects (HTTP calls, Telegram messages) **inside** SQL transactions:
  - use patterns like outbox / background workers instead.
- Migrations:
  - live in the `migrations/` directory;
  - must be ordered (timestamps or incremental numbers);
  - should be idempotent and reversible where supported.

---

## 7. Refactoring constraints

- Do **not** perform large, cross-cutting refactors without explicit request:
  - no mass renaming of packages/modules;
  - no breaking changes to public interfaces unless necessary.
- If refactoring is required:
  - describe the plan first;
  - outline risks and potential impact;
  - split into several micro-steps where possible.

---

## 8. Communication style

- Responses should be concise and practical.
- Code must be production-oriented:
  - no temporary hacks or commented-out junk unless requested for illustration.
- If the task is ambiguous:
  - ask for clarification before changing files where possible;
  - do not exceed the scope of the current micro-step.

---

## 9. Prohibited actions

- Do **not**:
  - change the primary project language (Go);
  - switch DB/cache/brokers without approval;
  - introduce heavy dependencies without clear justification;
  - ignore explicit constraints from the project specification or user instructions.
- Never commit or propose:
  - private keys,
  - real tokens/passwords,
  - any other sensitive data.

---

## 10. MCP usage policy (for Cursor / MCP-enabled environments)

If MCP providers (filesystem, shell, git, etc.) are available, the assistant **must prefer MCP tools** over guessing or free-text pseudo-commands.

### 10.1. General MCP rules

- **Plan, then execute.**
  - First: show a short *PLAN* (what files will be written, what commands run).
  - Then: use MCP tools to actually perform those actions.
- Prefer safe options:
  - use `--check` / `--diff` / dry-run when available (linters, formatters, migrations).
- No destructive commands (e.g. `rm -rf`, dropping databases, deleting large directories) unless the user explicitly asks for them in the current step.

### 10.2. Files via MCP

When modifying a file using MCP filesystem tools:

- Always write the **full new content** of the file.
- Ensure the content follows all rules in this document.

### 10.3. Commands via MCP shell

When running commands with shell MCP:

- Use only the commands that are relevant for build/test/lint/run/devops tasks.
- If a command fails:
  - show the exact error output;
  - explain the cause if possible;
  - propose a minimal fix;
  - only then suggest re-running the command.

### 10.4. Example MCP interaction format

When describing actions to take:

```text
PLAN:
- Overwrite: PROJECT_RULES.md
- Run: go fmt ./...
- Run: go build ./cmd/bot

EXECUTION (via MCP/Fs/Shell):
- fs.write PROJECT_RULES.md (full content)
- shell: go fmt ./...
- shell: go build ./cmd/bot
11. Priority of rules
If these rules conflict with an explicit instruction from the user in the current request,
the current user instruction takes priority.
