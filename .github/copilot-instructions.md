# GitHub Copilot Instructions for FM

For full project conventions, see AGENTS.md in the repository root.

## PR Review

When reviewing pull requests, do not flag the following patterns as issues.
Each is an intentional project convention:

- **CLAUDE.md is a symlink**: CLAUDE.md is a symlink to AGENTS.md, not a duplicated file. The symlink target content appearing in diffs is expected behavior. Do not flag it as duplication.
- **`.claude/settings.json` enabledPlugins is intentional**: The `enabledPlugins` field in `.claude/settings.json` configures project-level Claude Code plugins used for development. Do not suggest removing it or moving it to local-only config.
- **Plan documents are historical records**: Files in `docs/plans/` are implementation plans, not specifications. They document the rationale behind decisions and track follow-up work in dedicated sections. Do not suggest expanding plan documents with additional caveats or migration-window edge cases already addressed in their follow-up sections.
