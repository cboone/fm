# Add --subject filter to list and summary commands

## Context

Issue #72 requests a `--subject` filter flag across all query-oriented commands.
The action commands (archive, mark-read, unflag, spam, flag, move) and `search`
already have `--subject` via `addFilterFlags()` and inline flag registration
respectively. The remaining gap is `list` and `summary`, which currently only
support `--mailbox`, `--unread`, `--flagged`, `--unflagged`, and their own
specific flags.

**Classification:** feature
**Commit prefix:** `feat`

## Changes

### 1. Add `Subject` field to `ListOptions` (`internal/client/email.go:102`)

Add `Subject string` to the struct. In `ListEmails()` (~line 127), set
`fc.Subject = opts.Subject` when non-empty, matching the pattern used in
`buildSearchFilter()`.

### 2. Add `Subject` field to `SummaryOptions` (`internal/client/email.go:709`)

Add `Subject string` to the struct. In `AggregateSummary()` (~line 725), set
`fc.Subject = opts.Subject` when non-empty.

### 3. Wire `--subject` flag in `list` command (`cmd/list.go`)

- Register `cmd.Flags().String("subject", "", "filter by subject text")` in `init()`
- Read the flag in `RunE` and pass to `ListOptions.Subject`

### 4. Wire `--subject` flag in `summary` command (`cmd/summary.go`)

- Register `cmd.Flags().String("subject", "", "filter by subject text")` in `init()`
- Read the flag in `RunE` and pass to `SummaryOptions.Subject`

### 5. Add tests (`internal/client/email_test.go`)

- `TestListEmails_SubjectFilter`: Verify `fc.Subject` is set on the JMAP query
  filter. Follow the pattern from `TestListEmails_FlaggedOnly` (~line 1580).
- `TestAggregateSummary_SubjectFilter`: Verify `fc.Subject` is set on the JMAP
  query filter. Follow the pattern from `TestAggregateSummary_Basic` (~line 2692).

### 6. Update CLI reference (`docs/CLI-REFERENCE.md`)

- Add `--subject` row to the `list` flag table (~line 113)
- Add `--subject` row to the `summary` flag table (~line 540)
- Add filtering examples showing `--subject` usage for both commands

## Verification

1. `go build ./...` compiles without errors
2. `go test ./...` passes all tests (existing and new)
3. `go vet ./...` reports no issues
4. Verify `fm list --help` and `fm summary --help` show the new `--subject` flag
