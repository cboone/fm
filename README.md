# jm

A safe, read-oriented CLI for JMAP email (Fastmail). Designed for use by
Claude Code to read, search, and triage email from the command line.

Sending and deleting email are structurally disallowed.

## Install

```bash
go install github.com/cboone/jm@latest
```

## Configuration

Set your Fastmail API token (Settings > Privacy & Security > Integrations > API tokens):

```bash
export JMAP_TOKEN="fmu1-..."
```

Or create `~/.config/jm/config.yaml`:

```yaml
token: "fmu1-..."
```

## Commands

```
jm session                           # verify connectivity and auth
jm mailboxes                         # list all mailboxes
jm list                              # list emails in inbox
jm list --mailbox archive --limit 10 # list emails in archive
jm read <email-id>                   # read full email content
jm read <email-id> --thread          # read email with thread context
jm search "meeting agenda"           # full-text search
jm search --from alice@example.com   # search by sender
jm archive <email-id>                # move to archive
jm spam <email-id>                   # move to junk
jm move <email-id> --to Receipts     # move to a named mailbox
```

## Safety

- No sending email (no `EmailSubmission`)
- No deleting email (no `Email/set` destroy, no move to Trash)
- The `move` command refuses Trash, Deleted Items, and Deleted Messages

## Development

```bash
make          # build the binary
make test     # run unit tests
make test-cli # build + run scrut CLI tests
make test-all # run all tests (unit + CLI)
make help     # show all targets
```

See [PLAN.md](PLAN.md) for the full design document.
