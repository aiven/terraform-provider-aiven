# AI Agent Skills

Reusable skill definitions for AI coding assistants. Each skill provides structured instructions for common development tasks in this repository.

Repo-wide conventions (tooling, where new code goes) live in [`AGENTS.md`](../../../AGENTS.md) at the repo root and are auto-loaded by most agents. The skills here are the deep references for resource generation and migration specifically.

## Available Skills

| Skill | Description |
|-------|-------------|
| [tf-resource-generator](./tf-resource-generator/SKILL.md) | Generate new Terraform resources and data sources from YAML definitions and OpenAPI specs |
| [tf-resource-migration](./tf-resource-migration/SKILL.md) | Migrate existing SDK-based resources to the Plugin Framework while preserving state and behavior |

## Usage

### Claude Code

Copy the skill folder to `.claude/skills/` in your workspace:

```bash
cp -r tools/agents/skills/tf-resource-generator .claude/skills/
```

Then invoke with `/tf-resource-generator` or let Claude auto-detect when relevant.

**Docs:** [Claude Code Skills](https://docs.anthropic.com/en/docs/claude-code/skills)

### Cursor

Add skill content to `.cursor/rules/` or reference in `.cursorrules`:

```bash
cp tools/agents/skills/tf-resource-generator/SKILL.md .cursor/rules/tf-resource-generator.md
```

**Docs:** [Cursor Rules](https://docs.cursor.com/context/rules-for-ai)

### Other Agents

Most AI coding assistants support custom instructions. Copy the `SKILL.md` content to your agent's instruction/rules file.

## Contributing

When adding a new skill:
1. Create a folder: `tools/agents/skills/<skill-name>/`
2. Add `SKILL.md` with frontmatter and instructions
3. Update this README's "Available Skills" table
