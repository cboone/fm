# fm

`fm` is a safety-first email CLI for LLM agents.

It gives agents a reliable way to read, search, triage, and draft Fastmail email via JMAP, with structured output and hard limits on risky actions.

This tool is primarily designed for agent-driven workflows, not manual mailbox management.

## What Agents Can Do

- Read mailbox and message data (`session`, `mailboxes`, `list`, `read`, `search`)
- Build inbox intelligence (`stats`, `summary`)
- Run controlled triage actions (`archive`, `spam`, `mark-read`, `flag`, `unflag`, `move`)
- Create drafts for human review (`draft`)

## Hard Safety Model

`fm` enforces safety constraints in code:

- **No sending email:** `EmailSubmission` is never called
- **No deletion API calls:** `Email/set` destroy is never used
- **No move-to-trash loophole:** `move` refuses Trash, Deleted Items, and Deleted Messages targets
- **Draft-only composition:** `draft` creates in Drafts with `$draft`; it cannot send

## Install

### Homebrew

```bash
brew install cboone/tap/fm
```

### Go

```bash
go install github.com/cboone/fm@latest
```

## Agent Setup

1. Create a Fastmail API token at **Settings > Privacy & Security > Integrations > API tokens**.
2. Grant only:
   - `urn:ietf:params:jmap:core`
   - `urn:ietf:params:jmap:mail`
3. Set credentials in environment variables:

```bash
export FM_TOKEN="fmu1-..."
```

4. Validate connectivity:

```bash
fm session --format json
```

`urn:ietf:params:jmap:submission` is intentionally not required.

## Recommended Agent Workflow

If you are an LLM agent, use this loop:

1. **Discover context**: `fm session`, `fm mailboxes`, `fm summary --unread`
2. **Inspect candidates**: `fm list` and `fm read` (or `fm search` with filters)
3. **Preview bulk actions**: run triage commands with `--dry-run` first
4. **Apply changes**: execute the same command without `--dry-run`
5. **Verify state**: re-check with `fm summary`, `fm list`, or targeted `fm search`
6. **Draft replies only when needed**: create with `fm draft`; do not attempt to send

For bulk triage commands, provide either explicit email IDs or filter flags in a given call, not both.

## Agent-Friendly Command Patterns

```bash
# Baseline mailbox state
fm summary --unread --newsletters --format json

# Filter-first discovery
fm search --from billing@example.com --after 2026-01-01 --format json

# Preview then apply a safe bulk action
fm archive --from updates@example.com --dry-run --format json
fm archive --from updates@example.com --format json

# Flag color workflow
fm flag --color orange <email-id> --format json
fm unflag --color <email-id> --format json

# Draft creation (saved only)
fm draft --reply-to <email-id> --body "Thanks, will review today." --format json
```

All triage commands support `--dry-run` (`archive`, `spam`, `mark-read`, `flag`, `unflag`, `move`).

## Output Contract for Agents

- Default output is structured `json`
- `text` output exists for human readability (`--format text`)
- Runtime errors are structured on stderr (`json` or `text`) and return exit code `1`
- `partial_failure` means some IDs succeeded and some failed, parse stdout result and stderr error together

For full output schemas and error references, see [docs/CLI-REFERENCE.md](docs/CLI-REFERENCE.md).

## Configuration

Resolution order (highest first):

1. Command flags (`--token`, `--format`, etc.)
2. Environment variables (`FM_TOKEN`, `FM_FORMAT`, etc.)
3. Config file (`~/.config/fm/config.yaml`)

### Environment Variables

| Variable         | Description                     | Default                                 |
| ---------------- | ------------------------------- | --------------------------------------- |
| `FM_TOKEN`       | Bearer token for authentication | (none)                                  |
| `FM_SESSION_URL` | JMAP session endpoint           | `https://api.fastmail.com/jmap/session` |
| `FM_FORMAT`      | Output format: `json` or `text` | `json`                                  |
| `FM_ACCOUNT_ID`  | JMAP account ID override        | (auto-detected)                         |

### Optional Config File

```yaml
# ~/.config/fm/config.yaml
session_url: "https://api.fastmail.com/jmap/session"
format: "json"
account_id: ""
```

Security recommendation: keep tokens in environment variables, not in committed files.

## Claude Code Notes

`fm` works with any shell-capable agent runtime, including Claude Code.

This repository ships a Claude Code plugin with a `review-email` skill for guided multi-phase inbox triage.

Install from inside `claude`:

```text
/plugin marketplace add cboone/fm
```

After installation, prompts like "review my email", "triage email", and "check my inbox" can activate the skill workflow.

For Claude-specific setup and examples, see [docs/CLAUDE-CODE-GUIDE.md](docs/CLAUDE-CODE-GUIDE.md).

## Reference Docs

- Full command and schema reference: [docs/CLI-REFERENCE.md](docs/CLI-REFERENCE.md)
- Claude Code integration guide: [docs/CLAUDE-CODE-GUIDE.md](docs/CLAUDE-CODE-GUIDE.md)

## License

[MIT License](./LICENSE). Use freely, keep the notice, no warranty.
