# 2026-02-10: Improve text column formatting for mailboxes and email lists

Addresses: GitHub issues #3 and #4.

## Context

The text formatter in `internal/output/text.go` uses hardcoded fixed-width columns that break alignment or silently truncate content:

- **Mailboxes** (`formatMailboxes`, line 68): `%-40s` for mailbox name -- breaks alignment when names exceed 40 characters.
- **Email lists** (`formatEmailList`, line 85): `%-30s` for sender, `%-50s` for subject -- fixed minimum widths that cause uneven spacing and poor readability with longer values.

Both were identified during Copilot PR review of PR #1.

## Approach

### 1. Mailboxes: use `text/tabwriter`

Replace the fixed `%-40s` format string with tab-separated columns and `tabwriter` for dynamic alignment. This is a straightforward case since each mailbox is a single output line.

```go
tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
// write tab-separated columns
tw.Flush()
```

### 2. Email lists: first-pass width calculation

`tabwriter` is a poor fit here because each email produces multi-line output (main line + ID line + optional snippet). Instead:

1. **First pass**: iterate all emails, build display strings with truncation applied, track max sender/subject widths.
2. **Second pass**: use a dynamically computed `fmt.Sprintf("%%-%ds", maxWidth)` format string for aligned output.

This gives dynamic column alignment without tabwriter complications on continuation lines.

### 3. Add `truncate` helper with ellipsis

Even with dynamic widths, unbounded fields make output unwieldy. Add a rune-aware `truncate(s, maxLen)` helper that appends `"..."` when truncation occurs.

Constants for maximum column widths:
- `maxFromWidth = 40` (sender column)
- `maxSubjectWidth = 80` (subject column)

The helper uses `[]rune` conversion for safe truncation of multi-byte characters.

## Files to modify

| File | Changes |
|------|---------|
| `internal/output/text.go` | Add `"text/tabwriter"` import; add `truncate` helper + constants; rewrite `formatMailboxes` and `formatEmailList` |
| `internal/output/text_test.go` | Add tests for `truncate`, mailbox alignment, email list truncation, and email list alignment |

No changes to `formatter.go`, `types.go`, `json.go`, or any command files.

## Implementation steps

1. Add `truncate` helper function and width constants to `text.go`.
2. Rewrite `formatMailboxes` with `tabwriter`.
3. Rewrite `formatEmailList` with first-pass width calculation and truncation.
4. Add new tests: `TestTruncate`, `TestTextFormatter_MailboxesAlignment`, `TestTextFormatter_EmailListTruncation`, `TestTextFormatter_EmailListAlignment`.
5. Run `go test ./internal/output/` and confirm all existing + new tests pass.

## Verification

1. `go test ./internal/output/ -v` -- all tests pass.
2. `go vet ./...` -- no warnings.
3. Manual: `jm mailboxes --format text` with real data to visually confirm alignment.
4. Manual: `jm list --format text` to confirm sender/subject truncation with ellipsis and column alignment.
