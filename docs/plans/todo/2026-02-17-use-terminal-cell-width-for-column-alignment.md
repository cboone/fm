# Use Terminal Cell-Width for Column Alignment

## Context

The text formatter (`internal/output/text.go`) uses `utf8.RuneCountInString` for column width calculation and `fmt.Sprintf("%%-%ds", ...)` for padding. Both operate on Unicode code points (runes), not terminal display columns. This breaks column alignment when sender names or subjects contain characters that occupy more than one terminal cell, such as CJK ideographs (2 columns each), emoji, or zero-width combining marks.

This was identified during PR #14 Copilot review. The rune-based approach was an intentional improvement over byte-based alignment; this is the follow-up to make it fully correct. Tracked as issue #15.

## Approach

Add `github.com/mattn/go-runewidth` as a direct dependency and replace all rune-count-based width calculations and padding with cell-width-aware equivalents.

`go-runewidth` is the de facto standard Go library for terminal column width. It provides `StringWidth` (measure display columns), `FillRight` (pad to display width), and `Truncate` (truncate to display width) -- exactly what we need.

## Changes

### 1. Add dependency

```
go get github.com/mattn/go-runewidth
```

### 2. Update `internal/output/text.go`

**Imports**: Replace `unicode/utf8` with `github.com/mattn/go-runewidth`.

**Constants** (lines 15-18): Update comments from "rune count" to "display columns".

**`formatEmailList`** (lines 116-127):
- Replace `utf8.RuneCountInString(from)` and `utf8.RuneCountInString(subject)` with `runewidth.StringWidth(...)`.
- Replace the `fmt.Sprintf("%%-%ds", ...)` format string with direct `runewidth.FillRight` calls per column.

Before:
```go
fmtStr := fmt.Sprintf("%%s %%-%ds  %%-%ds  %%s\n", maxFrom, maxSubject)
fmt.Fprintf(w, fmtStr, r.unread, r.from, r.subject, r.date)
```

After:
```go
fmt.Fprintf(w, "%s %s  %s  %s\n", r.unread,
    runewidth.FillRight(r.from, maxFrom),
    runewidth.FillRight(r.subject, maxSubject),
    r.date)
```

**`formatDryRunResult`** (lines 245-256): Same pattern -- replace `utf8.RuneCountInString` with `runewidth.StringWidth`, replace `fmt.Sprintf` format string with `runewidth.FillRight`.

**`truncate`** (lines 319-327): Replace rune-based logic with `runewidth.StringWidth` for the check and `runewidth.Truncate` for the actual truncation.

Before:
```go
func truncate(s string, maxLen int) string {
    if maxLen < 4 || utf8.RuneCountInString(s) <= maxLen {
        return s
    }
    runes := []rune(s)
    return string(runes[:maxLen-3]) + "..."
}
```

After:
```go
func truncate(s string, maxWidth int) string {
    if maxWidth < 4 || runewidth.StringWidth(s) <= maxWidth {
        return s
    }
    return runewidth.Truncate(s, maxWidth, "...")
}
```

### 3. Update `internal/output/text_test.go`

**`TestTextFormatter_EmailListAlignmentMultiByteRunes`** (lines 876-921): Rewrite to verify display-column alignment by comparing two emails (one CJK, one ASCII) and checking their date columns align at the same display-column position using `runewidth.StringWidth`.

**`TestTruncate`**: Add test cases for wide characters to verify cell-width-aware truncation.

### Files modified

- `go.mod` / `go.sum` (new dependency)
- `internal/output/text.go` (width calculation, padding, truncation)
- `internal/output/text_test.go` (update multi-byte test, add wide-char truncation cases)

### Not changed

- `formatMailboxes` uses `text/tabwriter`, which has its own byte-based width issues with CJK. That is a separate concern (tabwriter would need wrapping or replacement) and is out of scope.
- `formatStats` only right-aligns integer counts (`%*d`), which are ASCII-only. No change needed.

## Verification

1. `go build ./...` -- compiles cleanly
2. `go test ./internal/output/...` -- all tests pass, including updated multi-byte alignment test
3. `go vet ./...` -- no issues
4. Run project linters via `lint-and-fix`
