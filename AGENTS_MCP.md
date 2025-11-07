# Himera Trading Bot – MCP Instructions for LLM Assistants

This document defines how any MCP-enabled LLM (Cursor, VS Code agents, etc.) must work in this repository.

## 1. Initial reading order

When you start working in this project, you MUST:

1. Read `PROJECT_RULES.md` from the repository root.
2. Read `README.md` from the repository root to understand:
   - architecture;
   - setup and tooling;
   - how to build, test, and run the bot.
3. Read the workspace MCP configuration:
   - preferred: `.vscode/settings.json` `"mcpServers"` section (if present);
   - optionally: user-level editor config if explicitly requested by the user.

You must not "guess" the project structure or tools. Use the real files via the filesystem MCP server.

## 2. Required MCP servers and their usage

For this project you MUST, whenever available, use:

- **filesystem MCP server**
  - Read and write real files under the project root.
  - Always rewrite entire files – no partial edits, no ellipsis.
- **shell / command MCP server** (if configured by the user)
  - Run only project-relevant commands (Go build/test/lint, scripts).
  - Example commands:
    - `go build ./cmd/bot`
    - `go test ./...`
    - `golangci-lint run ./...`
    - `pre-commit run --all-files`
- **github MCP server**
  - Only when explicitly asked to inspect remote repositories, PRs, or issues.
  - Do NOT modify this repository via GitHub MCP – local workspace is the source of truth.
- **web_browsing / HTTP tools**
  - Only for external API docs / library documentation when user requests it or when you need authoritative reference.

If a required MCP server is not available, you must explicitly say so to the user and fall back to normal reasoning.

## 3. Project root and file handling rules

- Project root: `/root/himera-bot` (unless the user tells you otherwise).
- Always treat this directory as the root for filesystem operations.
- When modifying any file:
  - Read the current content via filesystem MCP.
  - Generate the full new content (no `...`, no partial snippets).
  - Write it back via filesystem MCP.

## 4. Step-by-step workflow per micro-step

For each micro-step (e.g. `Phase1-M003`):

1. **PLAN**
   - List 2–7 concrete actions:
     - which files you will read;
     - which files you will create or fully rewrite;
     - which MCP servers you will use (`filesystem`, `shell`, `github`, etc.);
     - which commands you will run via `shell`.
2. **EXECUTION**
   - Use `filesystem` MCP to read current files.
   - Use `filesystem` MCP to write full updated files.
   - Use `shell` MCP to run Go build/test/lint commands requested by the user.
3. **REPORT**
   - Show the updated files in full.
   - Show command outputs that are relevant (build/test/lint results).
   - Do NOT silently change tooling (Go version, linters, Makefile, go.mod, etc.).

## 5. Respect existing project rules

This document extends, but does not override, `PROJECT_RULES.md`.

If anything here conflicts with an explicit instruction from the user in the current request, the user instruction has priority.
