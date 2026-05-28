# Repository Conventions for AI Coding Agents

This file is read automatically by Cursor, Claude Code, OpenAI Codex, GitHub Copilot, Aider, and other agents that support the [agents.md](https://agents.md) standard. Follow it.

## Tooling: always use Task

All build, generate, format, and lint workflows in this repo go through [Task](https://taskfile.dev). The task definitions live in [Taskfile.dist.yml](Taskfile.dist.yml) (plus `TaskInternal.yml`). Run `task --list` to discover commands, or read `Taskfile.dist.yml` directly.

The canonical commands you will need most:

- `task generate` — regenerates code from `definitions/*.yml` and the OpenAPI spec. It already runs `task fmt` (fast mode) at the end, so you do **not** need to format separately after generating.
- `task fmt` — formats Go, Terraform, and whitespace. Use after hand-written code changes.
- `task lint` — runs all linters (Go, tests, docs, semgrep). Run before considering work done.
- `task build` — compiles the provider. Use as a quick sanity check.
- `task test-unit` — unit tests (no API calls).

### Do NOT use

For the workflows above, do not reach for:

- `go run ./generators/...` or `go generate ./...` — use `task generate`.
- `gofmt`, `goimports`, `golangci-lint` invoked directly — use `task fmt` / `task lint`.
- `make` — this repo uses Task, not Make. There is no `Makefile`.

These bypass the project's wrappers (env vars, ordering, defer-fmt, multi-formatter pipeline) and will produce inconsistent results.

## Where new code goes

- New Terraform resources and data sources go in `internal/plugin/` (Plugin Framework, YAML-generated). Do not add new resources to `internal/sdkprovider/`.
- Generated files have the `zz_` prefix. Never edit them by hand — change the YAML in `definitions/` and run `task generate`.
- Hand-written custom logic (modifiers, view overrides) lives in non-`zz_` `.go` files in the same package.

## Deeper references

For end-to-end workflows, read the skills in [tools/agents/skills/](tools/agents/skills/README.md):

- [tf-resource-generator](tools/agents/skills/tf-resource-generator/SKILL.md) — creating new resources/data sources from YAML.
- [tf-resource-migration](tools/agents/skills/tf-resource-migration/SKILL.md) — migrating SDK resources to the Plugin Framework.

These skills are the source of truth for YAML schema, adapter API, modifier patterns, and state-compatibility rules. This file is the source of truth for project-wide tooling conventions.
