# Compact text output for action commands (#71)

## Context

Action commands (`archive`, `spam`, `flag`, `unflag`, `mark-read`, `move`) produce verbose output that lists every processed email ID. Even with `--format text`, the output includes all IDs, which is unusable for bulk operations (e.g., 527 IDs from a single archive command). During a triage session processing ~1,800 emails across 60+ senders, every action command had to be piped through Python just to extract counts.

The fix: make the text formatter produce a compact one-line summary for `MoveResult`, while keeping JSON output unchanged for programmatic use.

## Approach

Modify the existing `TextFormatter.formatMoveResult` in `internal/output/text.go` to produce a compact summary instead of listing every ID. No new format type, no changes to the `Formatter` interface, no changes to `MoveResult` struct or action command files.

### Output format

**No destination (flag, unflag, mark-read):**
```
Flagged 2 of 2 matched emails (0 failed)
```

**With destination, move only:**
```
Moved 5 of 5 matched emails to Receipts (0 failed)
```

**With destination, archive/spam (verb implies destination):**
```
Archived 527 of 527 matched emails (0 failed)
```

**Partial failure:**
```
Archived 526 of 527 matched emails (1 failed)
Errors:
  - M42: not found
```

### Verb mapping

Determined from which `MoveResult` slice field is populated:

| Field          | Verb             |
| -------------- | ---------------- |
| `Archived`     | "Archived"       |
| `MarkedSpam`   | "Marked as spam" |
| `MarkedAsRead` | "Marked as read" |
| `Flagged`      | "Flagged"        |
| `Unflagged`    | "Unflagged"      |
| `Moved`        | "Moved"          |
| (none/default) | "Processed"      |

The default case handles all-failed scenarios where no action slice is populated.

## Changes

### 1. `internal/output/text.go` (lines 211-241)

- Add private `actionVerb(r types.MoveResult) (string, int)` helper before `formatMoveResult`
- Rewrite `formatMoveResult` to:
  1. Call `actionVerb` to get the verb and success count
  2. Print one summary line (with "to {Name}" suffix only for `Moved` with a destination)
  3. Print errors block if any

### 2. `internal/output/text_test.go` (lines 423-559)

- Update 6 existing `MoveResult` tests to expect compact output instead of ID lists
- Fix incomplete test fixtures (several tests omit `Matched`/`Processed` fields)
- Add `TestTextFormatter_MoveResultAllFailed` for the all-failed edge case
- Add `TestTextFormatter_MoveResultMovedWithDestination` for the move-specific "to Name" format

### 3. `docs/CLI-REFERENCE.md`

Update the "Text output" blocks for all 6 action commands (lines 676-682, 731-737, 783-788, 845-850, 906-911, 964-970) to show the compact format.

## Files not changed

- `internal/types/types.go`: `MoveResult` struct is unchanged. ID slices still populated and serialized to JSON.
- `internal/output/json.go`: JSON output retains full detail including all IDs.
- `internal/output/formatter.go`: No changes to the interface or factory.
- `cmd/root.go`: Format flag validation unchanged (still "json" or "text").
- `cmd/archive.go`, `cmd/spam.go`, `cmd/flag.go`, `cmd/unflag.go`, `cmd/mark-read.go`, `cmd/move.go`: No changes. All continue to populate `MoveResult` the same way.

## Verification

1. `go build ./...` compiles without errors
2. `go test ./...` passes all tests
3. `go vet ./...` reports no issues
4. Run project linters (`lint-and-fix`)
