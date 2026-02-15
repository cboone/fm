# Add dry-run mode to mutating commands

Closes #25.

## Context

There is no safe preview mode for bulk actions. Users cannot confirm what will be modified before executing `archive`, `mark-read`, `flag`, `unflag`, `spam`, or `move`. This plan adds `--dry-run` (`-n`) to all six mutating commands so users can see a preview of affected emails without making server-side changes.

## Changes

### 1. Add `DryRunResult` type (`internal/types/types.go`)

Add after `MoveResult`:

```go
type DryRunResult struct {
    Operation   string           `json:"operation"`
    Count       int              `json:"count"`
    Emails      []EmailSummary   `json:"emails"`
    NotFound    []string         `json:"not_found,omitempty"`
    Destination *DestinationInfo `json:"destination,omitempty"`
}
```

`Operation` is one of: `archive`, `move`, `spam`, `mark_read`, `flag`, `unflag`. `Emails` reuses the existing `EmailSummary` type. `NotFound` captures IDs that fail `Email/get` (e.g. already deleted). `Destination` is populated for archive/spam/move.

### 2. Add `GetEmailSummaries` client method (`internal/client/email.go`)

New method that does a read-only `Email/get` with `summaryProperties` (line 30) for a list of IDs, batched in groups of `batchSize` (50). Returns `([]types.EmailSummary, []string, error)` -- found summaries, not-found IDs, and error.

Reuses the existing `convertSummaries` helper (line 742) and `summaryProperties` slice (line 30). Never calls `Email/set`.

### 3. Add dry-run helper (`cmd/dryrun.go`)

New file with a single function:

```go
func dryRunPreview(c *client.Client, ids []string, operation string, dest *types.DestinationInfo) error
```

Calls `c.GetEmailSummaries(ids)`, builds a `DryRunResult`, outputs via `formatter().Format(os.Stdout, result)`.

### 4. Add `--dry-run` flag and branching to each mutating command

For each of `archive.go`, `spam.go`, `move.go`, `mark-read.go`, `flag.go`, `unflag.go`:

- Register `--dry-run` (`-n`) bool flag in `init()`
- In `RunE`, after authentication and mailbox resolution (where applicable) but *before* the mutation call, check the flag and early-return with `dryRunPreview()`

This means dry-run still validates credentials and target mailbox. It only skips the `Email/set` mutation.

For commands with a destination (archive, spam, move): pass `&types.DestinationInfo{...}`.
For commands without (mark-read, flag, unflag): pass `nil`.

Operation strings: `"archive"`, `"spam"`, `"move"`, `"mark_read"`, `"flag"`, `"unflag"`.

### 5. Add text formatting for `DryRunResult` (`internal/output/text.go`)

Add `types.DryRunResult` case to the `Format()` type switch and a `formatDryRunResult` method.

Text output format:

```
Dry run: would archive 3 email(s)

  M1  Alice <alice@example.com>  Meeting tomorrow  2026-02-14 10:30
  M2  Bob <bob@example.com>      Invoice #1234     2026-02-13 09:15

Destination: Archive (mb-archive-id)
```

If `NotFound` is non-empty, append: `Not found: M4, M5`

JSON output requires no changes -- `JSONFormatter` handles any struct via `encoding/json`.

### 6. Add tests

**`internal/client/email_test.go`**:
- `GetEmailSummaries` returns correct summaries for found IDs
- `GetEmailSummaries` returns not-found IDs separately
- `GetEmailSummaries` handles `jmap.MethodError`
- `GetEmailSummaries` batches correctly (51 IDs = 2 `Do` calls)

**`internal/output/text_test.go`** (or `internal/output/formatter_test.go`):
- Text formatting of `DryRunResult` with destination
- Text formatting without destination (mark-read/flag/unflag)
- Text formatting with not-found IDs

**`tests/help.md`**:
- Update help expectations for each mutating command to include `--dry-run`

**`tests/flags.md`**:
- Verify `--dry-run` is accepted on each mutating command (gets auth error, not flag error)

### 7. Update documentation (`docs/CLI-REFERENCE.md`)

- Add `--dry-run` / `-n` to the flags table for each mutating command
- Add `DryRunResult` to the Output Schemas section

## Files to modify

| File | Change |
|------|--------|
| `internal/types/types.go` | Add `DryRunResult` struct |
| `internal/client/email.go` | Add `GetEmailSummaries` method (reuses `summaryProperties`, `convertSummaries`) |
| `cmd/dryrun.go` | **New file** -- `dryRunPreview` helper |
| `cmd/archive.go` | Add `--dry-run` flag + early-return branch |
| `cmd/spam.go` | Add `--dry-run` flag + early-return branch |
| `cmd/move.go` | Add `--dry-run` flag + early-return branch |
| `cmd/mark-read.go` | Add `--dry-run` flag + early-return branch |
| `cmd/flag.go` | Add `--dry-run` flag + early-return branch |
| `cmd/unflag.go` | Add `--dry-run` flag + early-return branch |
| `internal/output/text.go` | Add `formatDryRunResult` method + type-switch case |
| `internal/client/email_test.go` | Tests for `GetEmailSummaries` |
| `internal/output/text_test.go` | Tests for `formatDryRunResult` |
| `tests/help.md` | Update help expectations for mutating commands |
| `tests/flags.md` | Add `--dry-run` flag acceptance tests |
| `docs/CLI-REFERENCE.md` | Document `--dry-run` flag and `DryRunResult` schema |

## Verification

1. `go build ./...` compiles without errors
2. `go vet ./...` clean
3. `go test ./...` passes all existing and new tests
4. `make test-cli` -- scrut tests pass (help output, flag acceptance)
5. `jm archive --dry-run M1 M2` shows preview without mutating
6. `jm flag --dry-run M1` shows preview without mutating
7. `jm archive --dry-run --format text M1` renders human-readable text
8. `jm archive --help` shows `--dry-run` / `-n` flag
