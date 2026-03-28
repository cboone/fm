# fm

`fm` is a safety-first Fastmail CLI built for LLM agents. Use `fm` as the execution layer for inbox analysis and triage. It is machine-first by design: JSON output by default, structured errors, and strict guardrails around risky operations.

## Human Instructions (tl;dr)

Paste into your agent chat:

```text
Use `fm` for Fastmail triage. Read and follow the runbooks in `README.md` (section "Agent Runbook (Read First)") and `skills/review-email/references/runbook.md`.
```

## Agent Mission

Use `fm` to:

- Inspect mailbox state and message content
- Classify and batch-triage email safely
- Create drafts for human review

Do not use `fm` to send or delete email. Those operations are intentionally unavailable.

## Hard Safety Boundaries

`fm` enforces these guarantees in code:

- **No send path:** `EmailSubmission` is never called
- **No delete path:** `Email/set` destroy is never used
- **No trash-target moves:** `move` refuses Trash, Deleted Items, and Deleted Messages
- **Draft-only composition:** `draft` creates messages in Drafts with `$draft` and cannot send

Treat these as platform invariants, not optional settings.

## Agent Runbook (Read First)

If you are an LLM agent, follow this sequence on every run:

1. Prefer `json` output (`--format json` or default).
2. Start with read-only commands (`session`, `mailboxes`, `summary`, `list`, `search`, `read`).
3. Before any mutation, run the same command with `--dry-run`.
4. Do not mix explicit email IDs and filter flags in a single bulk command.
5. Apply mutations only after preview results look correct.
6. Verify post-state with follow-up reads (`summary`, `list`, `search`).
7. If the user asks to reply, create a draft and clearly state it was not sent.

## Quick Install

### Homebrew

```bash
brew install cboone/tap/fm
```

### Go

```bash
go install github.com/cboone/fm@latest
```

## Runtime Setup For Agents

1. Create a Fastmail API token at **Settings > Privacy & Security > Integrations > API tokens**.
2. Grant only these scopes:
   - `urn:ietf:params:jmap:core`
   - `urn:ietf:params:jmap:mail`
3. Store the token in your OS keychain:

**macOS (Keychain):**

```bash
security add-generic-password -s fm -a fastmail -w "fmu1-..."
```

**Linux (libsecret):**

```bash
echo -n "fmu1-..." | secret-tool store --label "fm" service fm
```

On macOS and Linux, `fm` retrieves the token from the OS keychain by default. No extra configuration is needed.

To use a different credential store, set a custom command:

```bash
export FM_CREDENTIAL_COMMAND="op read op://Private/Fastmail/token"
```

4. Validate auth and endpoint:

```bash
fm session
```

`urn:ietf:params:jmap:submission` is intentionally not required.

## Canonical Agent Loop

Use this loop for consistent, auditable behavior.

### 1) Discover

```bash
fm session
fm mailboxes
fm summary --unread --newsletters
```

### 2) Inspect

```bash
fm list --mailbox inbox --limit 50
fm search --from billing@example.com --after 2026-01-01
fm read <email-id>
```

### 3) Preview Mutations

```bash
fm archive --from updates@example.com --dry-run
fm mark-read --from updates@example.com --dry-run
fm move --from receipts@example.com --to Receipts --dry-run
```

### 4) Apply Mutations

```bash
fm archive --from updates@example.com
fm mark-read --from updates@example.com
fm move --from receipts@example.com --to Receipts
```

### 5) Verify

```bash
fm summary --unread
fm search --from updates@example.com --mailbox inbox
```

### 6) Draft If Needed

```bash
fm draft --reply-to <email-id> --body "Thanks, reviewed."
```

## Command Roles For Agents

| Role              | Commands                                                 |
| ----------------- | -------------------------------------------------------- |
| Auth and topology | `session`, `mailboxes`                                   |
| Discovery         | `list`, `search`                                         |
| Deep inspection   | `read`                                                   |
| Analytics         | `stats`, `summary`                                       |
| Triage mutations  | `archive`, `spam`, `mark-read`, `flag`, `unflag`, `move` |
| Draft composition | `draft`                                                  |
| Shell integration | `completion`                                             |

All triage mutations support `--dry-run`: `archive`, `spam`, `mark-read`, `flag`, `unflag`, `move`.

## Drafting Protocol

`draft` supports four modes: new, reply, reply-all, and forward.

```bash
fm draft --to alice@example.com --subject "Hello" --body "Hi Alice"
fm draft --reply-to <email-id> --body "Thanks!"
fm draft --reply-all <email-id> --body "Noted, thanks."
fm draft --forward <email-id> --to bob@example.com --body "FYI"
echo "Body text" | fm draft --to alice@example.com --subject "Hello" --body-stdin
```

Agent guidance:

- Draft when user intent is to respond or prepare a response
- Never claim the email was sent
- Remind the user draft review and send happen in Fastmail

## Output And Error Contract

- Default stdout format is structured `json`
- `--format text` is for human-readable output
- Runtime errors are structured on stderr and return exit code `1`
- `partial_failure` means mixed success. Parse both stdout and stderr.

Common error codes:

- `authentication_failed`
- `not_found`
- `forbidden_operation`
- `jmap_error`
- `network_error`
- `general_error`
- `config_error`
- `partial_failure`

For complete schemas and examples, see [docs/CLI-REFERENCE.md](docs/CLI-REFERENCE.md).

## Configuration

Resolution order (highest first):

1. Command flags (`--credential-command`, `--format`, etc.)
2. Environment variables (`FM_CREDENTIAL_COMMAND`, `FM_FORMAT`, etc.)
3. Config file (`~/.config/fm/config.yaml`)
4. Platform default (OS keychain on macOS and Linux)

### Environment Variables

| Variable                 | Description                                        | Default                                                |
| ------------------------ | -------------------------------------------------- | ------------------------------------------------------ |
| `FM_CREDENTIAL_COMMAND`  | Shell command that prints the API token to stdout   | macOS: OS keychain; Linux: libsecret; other: (none)    |
| `FM_SESSION_URL`         | JMAP session endpoint                              | `https://api.fastmail.com/jmap/session`                |
| `FM_FORMAT`              | Output format: `json` or `text`                    | `json`                                                 |
| `FM_ACCOUNT_ID`          | JMAP account ID override                           | (auto-detected)                                        |

### Optional Config File

```yaml
# ~/.config/fm/config.yaml
credential_command: "op read op://Private/Fastmail/token"
session_url: "https://api.fastmail.com/jmap/session"
format: "json"
account_id: ""
```

## Claude Code Specific Notes

`fm` works with any shell-capable agent runtime.

For Claude Code, this repository includes a plugin with a `review-email` skill for guided, multi-phase triage.

Install in `claude`:

```text
/plugin marketplace add cboone/fm
```

Common trigger prompts:

- "review my email"
- "triage email"
- "check my inbox"

Claude-specific guide: [docs/CLAUDE-CODE-GUIDE.md](docs/CLAUDE-CODE-GUIDE.md)

## Reference Docs

- Full command and schema reference: [docs/CLI-REFERENCE.md](docs/CLI-REFERENCE.md)
- Claude Code integration guide: [docs/CLAUDE-CODE-GUIDE.md](docs/CLAUDE-CODE-GUIDE.md)

## License

[MIT License](./LICENSE). TL;DR: Do whatever you want with this software, just keep the copyright notice included. The authors aren't liable if something goes wrong.
