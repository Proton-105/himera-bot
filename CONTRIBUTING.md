# Contributing to Himera Trading Bot

We welcome contributions that improve the Himera trading bot and its infrastructure. This guide explains how to work with the repository so your changes land smoothly.

## Branch strategy

- `main` is the stable branch. GitHub Actions deploys it automatically to staging/production once CI succeeds.
- Create topic branches from `main` using the following prefixes:
  - `feature/<short-name>` for new functionality
  - `fix/<short-name>` for bug fixes
  - `docs/<short-name>` for documentation-only changes
  - `chore/<short-name>` for maintenance and tooling updates
- Keep branches focused; open a Pull Request back to `main` as soon as the change is ready (or for review as a draft).

## Commit messages

Follow the [Conventional Commits](https://www.conventionalcommits.org/) format:

- `feat: ...`
- `fix: ...`
- `docs: ...`
- `chore: ...`
- `refactor: ...`
- `test: ...`

Always include a clear scope in parentheses (e.g., `bot`, `health`, `database`, `config`):

- `feat(bot): add /start handler`
- `fix(health): handle redis timeout`
- `chore(ci): update golangci-lint version`

## Code style (Go)

- Run `gofmt` on every file (enforced by pre-commit hooks, but still required).
- Package layout:
  - `cmd/bot` — application entrypoint and dependency wiring.
  - `internal/...` — domain entities, services, repositories, handlers, state management.
  - `pkg/...` — reusable infrastructure (configuration, logging, Redis adapters, etc.).
- Write identifiers and comments in English.
- Prefer returning `error`; avoid `panic` in business logic (panics are acceptable only in bootstrap code such as `main` when failing fast is necessary).

## Before you push

Run the local checklist to catch issues early:

```bash
go test ./...
golangci-lint run ./...
pre-commit run --all-files
go build ./cmd/bot        # when you touched build-related code
```

Pull Requests without a green CI (lint, tests, coverage, gosec, docker-build) will not be merged.

## Pull Request process

- Include in the PR description:
  - A concise summary of the change.
  - A link to the related ticket (Trello/Jira/GitHub issue).
  - Steps to test or verify the change.
- At least one reviewer must approve the PR.
- Do **not** push directly to `main`; every change goes through a PR.

## Environment & setup

For installation, configuration, and local run instructions, see [README.md](README.md).
