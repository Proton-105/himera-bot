# Himera Trading Bot – MCP Instructions for LLM Assistants

This document defines how any MCP-enabled LLM (Cursor, VS Code agents, Codex CLI, etc.) must work in this repository.

---

## 1. Initial reading order

When you start working in this project, you MUST:

1. Read `PROJECT_RULES.md` from the repository root.
2. Read `README.md` from the repository root to understand:
   - architecture;
   - setup and tooling;
   - how to build, test, and run the bot.
3. Read this file `AGENTS_MCP.md` to understand which MCP servers are available and how to use them.

You must not "guess" the project structure or tools. Use the real files via the **filesystem** MCP server.

---

## 2. Available MCP servers (names must match settings.json)

In the editor configuration (`settings.json`) the following MCP servers are available.
Use their **IDs exactly as listed here** when choosing tools:

- `filesystem`
  - Work with project files under the repo root.
  - Read/write Go code, configs, docs, migrations, etc.

- `sequential-thinking`
  - Helps with multi-step planning and breaking down complex tasks.
  - Use when you need to:
    - design micro-steps for a phase;
    - restructure a large refactor plan;
    - reason about tricky concurrency / state-machine flows.

- `context7`
  - Vector / semantic search over previously indexed content.
  - Use to:
    - find similar code/logic in the project;
    - locate related docs / examples by meaning, not exact text.

- `filesystem` (already mentioned, but key one)
  - MUST be used for all file operations in this repo.
  - Always rewrite whole files, no partial patches.

- `claude-context`
  - External Milvus/Zilliz-based context.
  - Use only if the user explicitly configured `OPENAI_API_KEY`, `MILVUS_ADDRESS`, `MILVUS_TOKEN`.
  - Optional; do not assume it’s configured.

- `supabase`
  - Remote HTTP MCP; use only if the user explicitly asks to interact with Supabase.
  - Not needed for core backend Go work unless explicitly requested.

- `tailwindcss`
  - TailwindCSS helper; only relevant for frontend / UI tasks.
  - For this Go/Telegram bot backend, usually NOT needed unless the user asks for UI/HTML.

- `shadcn`
  - shadcn/ui MCP; same as above — only for frontend-related tasks.

- `github`
  - Access GitHub APIs using the provided token.
  - Use to:
    - inspect remote repos, PRs, issues;
    - search in remote history if the user asks.
  - Do NOT modify this repo via GitHub MCP; the local workspace is the source of truth.

- `linear`
  - Work with Linear tasks/boards if the user wants you to sync issues/tasks.
  - Only use when explicitly asked.

- `ref_tools`
  - Generic reference tools (external info, docs).
  - Use if you need extra reference beyond web browsing and the user allows it.

- `web_browsing`
  - Browser-like MCP for external websites and documentation.
  - Use when you need up-to-date language/library/framework docs or when the user explicitly asks you to “look it up on the internet”.

If some of these servers are not actually available at runtime, you MUST say so to the user and gracefully fall back to normal reasoning.

---

## 3. Required MCP usage patterns

When MCP servers are available, you MUST prefer them over guessing:

- For any file inspection or modification in this repo:
  - Use `filesystem` MCP.
  - Always read the current version of the file first.
  - When changing a file, **rewrite the whole file** with consistent content (no `...`).

- For planning complex changes, multi-step refactors, or phase breakdown:
  - Use `sequential-thinking` MCP to create a step-by-step plan before editing files.

- For searching related code or docs by meaning:
  - Prefer `context7` (and/or `filesystem` search if available).

- For external docs / articles / API references:
  - Use `web_browsing` (and optionally `ref_tools` if appropriate).

- For GitHub-specific operations:
  - Use `github` MCP only when:
    - user asks about remote repo, PR, issue;
    - you need to see upstream history not present locally.

If a required MCP server fails or is unknown, explicitly mention which server failed and continue in the safest possible way.

---

## 4. Project root and file handling rules

- Project root: treat the repository root (where `PROJECT_RULES.md` and `README.md` live) as the base directory.
- When working with files via `filesystem` MCP:
  - Always use paths relative to the project root.
  - Do not access parent directories above the project root unless the user explicitly asks.

When modifying any file:

1. Read the current content via `filesystem`.
2. Generate the **full new content** of the file (no partial snippets, no `...`).
3. Write it back via `filesystem`.

---

## 5. Workflow per micro-step (Phase / M00X)

For each micro-step (e.g. `Phase1-M003`):

1. **PLAN**
   - List 2–7 concrete actions:
     - which files you will read;
     - which files you will create or fully rewrite;
     - which MCP servers you will use (`filesystem`, `sequential-thinking`, `context7`, `shell`, etc.);
     - which commands you will run via a shell-type MCP (if available).

2. **EXECUTION**
   - Use `filesystem` MCP to read current files.
   - Use `sequential-thinking` (optional but recommended) to refine the plan if the change is complex.
   - Use `filesystem` MCP to write full updated files.
   - Use a shell/command MCP (if configured) to run:
     - `go build ./cmd/bot`
     - `go test ./...`
     - `golangci-lint run ./...`
     - `pre-commit run --all-files`
     or other project-relevant commands requested by the user.

3. **REPORT**
   - Show the updated files in full.
   - Show relevant command outputs (build/test/lint results).
   - Do NOT silently change:
     - Go version;
     - toolchain / linters;
     - `go.mod`, `go.sum`, `Makefile` or CI config, unless the user explicitly asks.

---

## 6. Relationship to PROJECT_RULES.md

This document extends, but does NOT override, `PROJECT_RULES.md`.

- All constraints from `PROJECT_RULES.md` remain in force.
- If `AGENTS_MCP.md` and `PROJECT_RULES.md` conflict, assume `PROJECT_RULES.md` has priority unless the user clearly states otherwise.
- If both this file and `PROJECT_RULES.md` conflict with an explicit user instruction in the current request, **the current user instruction wins**.
instruction has priority.
