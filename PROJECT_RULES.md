# Himera Trading Bot — Project Rules for LLM Assistants (with MCP)

This document defines how LLM assistants (Cursor, Codex, ChatGPT, etc.) must operate in the **Himera** trading bot project.

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

## 10. MCP usage policy

### 10.1. Where MCP configuration lives

The assistant **must not guess** where MCP servers come from.

- In **editor clients** (Cursor, VS Code agents, etc.):
  - MCP servers are usually configured in editor settings like:
    - global: `~/.config/Code/User/settings.json`
    - workspace: `.vscode/settings.json`

- In **Codex CLI over SSH** (what you are using on the server):
  - Editor settings are **ignored**.
  - MCP configuration comes from a TOML file on the **remote machine**, e.g.:
    - `~/.codex/config.toml` (or whatever Codex CLI uses as its MCP config).
  - If Codex reports `unknown MCP server`, the assistant must assume:
    - either the server is not defined in this TOML,
    - or the name is misspelled in the prompt.

For this project, the assistant must assume that **Codex CLI reads MCP servers from the TOML config on the server** and not from VS Code `settings.json`.

### 10.2. How to use MCP (conceptually)

When MCP servers are available (as defined in the Codex TOML config), the assistant should:

- Use **filesystem-like MCP** to read and write real project files  
  instead of inventing their contents.
- Use **shell-like MCP** to run real commands:
  - `go build ./cmd/bot`
  - `go test ./...`
  - `golangci-lint run ./...`
  - `pre-commit run --all-files`
- Use any **planning / context MCP** (if configured) to:
  - build a short plan for complex micro-steps;
  - store/reuse long-term project decisions.
- Use any **web / HTTP MCP** only when:
  - local code and docs in the repo are not enough;
  - you explicitly need external documentation.

The exact server names and arguments are taken from the TOML MCP config;  
the assistant must not hard-code or assume them in this file.

### 10.3. MCP workflow per micro-step

For each micro-step in this repository, when MCP is available:

1. **Read rules first**
   - Read `PROJECT_RULES.md`.
   - Read `AGENTS_MCP.md` if the user mentions agents/MCP directly.

2. **Plan**
   - Produce a short `PLAN:` (2–7 bullets):
     - which files will be created/overwritten;
     - which commands will be run;
     - which MCP categories will be used (filesystem/shell/etc.).

3. **Execute via MCP**
   - Use filesystem-style MCP to:
     - read current file content;
     - write full new file content (no partial patches).
   - Use shell-style MCP to:
     - run build/test/lint commands given by the user.

4. **Report**
   - Show updated file contents in full (or only relevant ones, if there are many).
   - Show command outputs (build/test/lint), especially on failures.

### 10.4. Safety rules for MCP

- No destructive commands (`rm -rf`, dropping DBs, deleting large directories)  
  unless the user explicitly requests them **in this micro-step**.
- If a command fails:
  - show the exact error;
  - briefly explain what it means;
  - suggest a minimal, safe fix.

---

## 11. Priority of rules

If these rules conflict with an explicit instruction from the user in the current request,  
the current user instruction takes priority — **unless** it directly violates:

- security/secrets rules (sections 4 and 9);
- data safety (no destructive commands without explicit consent);
- obviously dangerous behavior (deleting code/data without confirmation).

In such cases, explain the conflict to the user and propose a safer alternative.  
Otherwise, follow the user’s current instruction and keep changes within the scope of the active micro-step.
