# Expose To and CC in search and list output

Addresses: #73

## Context

The `search` and `list` commands show `from`, `subject`, `received_at`, and `preview` but not the `to` address. During triage (e.g., building sender-to-identity mappings for unsubscribe drafts), the workaround requires fetching each email individually via `fm read`. Exposing `to` (and `cc`) in list/search output eliminates that round-trip.

**Current state:** `EmailSummary` already has a `To` field and the JMAP fetch already requests `"to"`. The JSON output already includes it. The gap is: (1) text output doesn't display `To`, and (2) `CC` is missing entirely from the summary pipeline.

## Changes

### 1. Add CC to `EmailSummary` struct

**File:** `internal/types/types.go` (line 33, after `To`)

Add:
```go
CC  []Address `json:"cc,omitempty"`
```

`omitempty` keeps JSON output clean when CC is empty.

### 2. Add "cc" to `summaryProperties`

**File:** `internal/client/email.go` (line 90)

Change `"from", "to",` to `"from", "to", "cc",` so JMAP returns CC for list/search results.

### 3. Map CC in `convertSummaries`

**File:** `internal/client/email.go` (line 1078, after `To`)

Add `CC: convertAddresses(e.CC),` to the `EmailSummary` literal.

### 4. Display To and CC in text list output

**File:** `internal/output/text.go` (lines 145-154, `formatEmailList` second pass)

Insert To and CC detail lines between the main row and the ID line:

```go
// After the main row fprintf...
if len(result.Emails[i].To) > 0 {
    _, _ = fmt.Fprintf(w, "  To: %s\n", formatAddrs(result.Emails[i].To))
}
if len(result.Emails[i].CC) > 0 {
    _, _ = fmt.Fprintf(w, "  CC: %s\n", formatAddrs(result.Emails[i].CC))
}
// Then ID line, then snippet...
```

Both are conditional (shown only when non-empty), matching the pattern in `formatEmailDetail`.

Resulting text output:
```
* Alice <alice@test.com>  Meeting tomorrow  2026-02-04 10:30
  To: Bob <bob@test.com>
  CC: Charlie <charlie@test.com>
  ID: M1
```

### 5. Update unit tests

**File:** `internal/output/text_test.go`

- `TestTextFormatter_EmailList`: add assertion for `To: Bob <bob@test.com>`
- `TestTextFormatter_EmailListWithSnippet`: add assertion for `To: me@test.com`
- New `TestTextFormatter_EmailListWithCC`: EmailSummary with CC populated, assert CC line appears
- New `TestTextFormatter_EmailListEmptyCC`: EmailSummary with nil CC, assert no `CC:` line
- New `TestTextFormatter_EmailListEmptyTo`: EmailSummary with nil To, assert no `To:` line

### 6. Update scrut live test

**File:** `tests/live.md` (lines 63-64)

The "List with text format includes Total and ID lines" test currently expects:
```
Total: * (glob)
* ID: * (glob)
```

Update to account for the new To line between the main row and ID:
```
Total: * (glob)
*To: * (glob)
* ID: * (glob)
```

## Files to modify

| File | Change |
|------|--------|
| `internal/types/types.go` | Add `CC` field to `EmailSummary` |
| `internal/client/email.go` | Add `"cc"` to `summaryProperties`; add `CC` in `convertSummaries` |
| `internal/output/text.go` | Add To/CC detail lines in `formatEmailList` |
| `internal/output/text_test.go` | Update existing tests, add new CC/empty tests |
| `tests/live.md` | Update glob pattern for list text output |

## What does NOT change

- **JSON output**: already works for To; CC is handled by adding the struct field with `json:"cc,omitempty"`
- **`summaryProperties` for To**: already includes `"to"`
- **`convertSummaries` for To**: already maps `e.To`
- **Command files** (`cmd/list.go`, `cmd/search.go`): pass results to formatter unchanged
- **`formatEmailDetail`**: already shows To and CC

## Verification

1. `go test ./internal/output/` -- existing + new tests pass
2. `go test ./internal/client/` -- no breakage from new CC field
3. `go vet ./...` and linter -- no issues
4. Manual: `go run . search --from "someone" --format text` shows To line; `--format json` includes `to` and `cc`
5. Scrut: `tests/live.md` glob patterns match updated output
