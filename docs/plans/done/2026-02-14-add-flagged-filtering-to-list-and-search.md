# Add flagged filtering to list and search

Closes #20.

## Context

`jm list` and `jm search` return `is_flagged` on each email but provide no way to filter by flag state. The `$flagged` keyword infrastructure already exists (flag/unflag commands, `IsFlagged` field in summaries/details, `HasKeyword`/`NotKeyword` fields in go-jmap's `FilterCondition`). This plan adds `--flagged` and `--unflagged` filter flags to both commands.

## Key constraint

go-jmap's `FilterCondition` has a single `NotKeyword` string field. The existing `--unread` filter uses `NotKeyword = "$seen"`, while `--unflagged` needs `NotKeyword = "$flagged"`. When both are active, a `FilterOperator{AND, [...]}` compound filter is required.

`HasKeyword` is also a single string field, but `--flagged` (`HasKeyword = "$flagged"`) never conflicts with `--unread` (`NotKeyword = "$seen"`), so it works in a simple `FilterCondition`.

## Changes

### 1. Add fields to `SearchOptions` (`internal/client/email.go:387`)

Add `FlaggedOnly bool` and `UnflaggedOnly bool` fields to the `SearchOptions` struct.

### 2. Update `SearchEmails` filter building (`internal/client/email.go:254-283`)

Use a shared `email.Filter` interface variable for the final filter so that both `Email/query` (line 296) and `SearchSnippet/get` (line 322) receive the same filter, whether it is a simple `FilterCondition` or a compound `FilterOperator`.

After building the `FilterCondition`:
- If `FlaggedOnly`: set `filter.HasKeyword = "$flagged"`
- If `UnflaggedOnly` and `!UnreadOnly`: set `filter.NotKeyword = "$flagged"`
- If `UnflaggedOnly` and `UnreadOnly`: the single `FilterCondition` already has `NotKeyword = "$seen"`. Create a second `FilterCondition{NotKeyword: "$flagged"}` and wrap both in a `FilterOperator{Operator: jmap.OperatorAND, Conditions: [filter, secondFilter]}`. Assign this compound filter to the shared variable so both `Email/query` and `SearchSnippet/get` use it.

### 3. Update `ListEmails` signature and filter building (`internal/client/email.go:42-55`)

Add `flaggedOnly bool` and `unflaggedOnly bool` parameters to `ListEmails`. Apply the same filter-building logic as step 2. In the compound case (`unflaggedOnly` + `unreadOnly`), keep `InMailbox` on the first `FilterCondition` inside the `FilterOperator` so the mailbox scope is preserved.

(Follow-up: refactor `ListEmails` to accept an options struct -- see #21.)

### 4. Add `--flagged` and `--unflagged` flags to `list` command (`cmd/list.go`)

- Register `--flagged` (`-f`) and `--unflagged` bool flags in `init()`
- Read both in `RunE`, validate they are not both set (return `general_error` if so)
- Pass to `ListEmails`

### 5. Add `--flagged` and `--unflagged` flags to `search` command (`cmd/search.go`)

- Register `--flagged` (`-f`) and `--unflagged` bool flags in `init()`
- Read both in `RunE`, validate they are not both set (return `general_error` if so)
- Set on `SearchOptions`

### 6. Add tests

**`cmd/list_test.go`** (or new `cmd/search_test.go`):
- Test that `--flagged` and `--unflagged` cannot both be set (validation)

**`internal/client/email_test.go`**:
- Test `SearchEmails` with `FlaggedOnly: true` sets `HasKeyword: "$flagged"` on the filter
- Test `SearchEmails` with `UnflaggedOnly: true` sets `NotKeyword: "$flagged"` on the filter
- Test `SearchEmails` with `UnflaggedOnly: true` + `UnreadOnly: true` produces a compound `FilterOperator` with AND
- Test that `FlaggedOnly` + `UnflaggedOnly` together is handled (validation at command level)
- Test `ListEmails` with `flaggedOnly` sets `HasKeyword: "$flagged"` on the filter
- Test `ListEmails` with `unflaggedOnly` sets `NotKeyword: "$flagged"` on the filter
- Test `ListEmails` with `unflaggedOnly` + `unreadOnly` produces a compound `FilterOperator` that preserves `InMailbox`

### 7. Update documentation (`docs/CLI-REFERENCE.md`)

- Add `--flagged` / `-f` and `--unflagged` rows to both the `list` and `search` flag tables
- Note that `--flagged` and `--unflagged` are mutually exclusive
- Add examples

## Files to modify

| File | Change |
|------|--------|
| `internal/client/email.go` | Add fields to `SearchOptions`, update `SearchEmails` and `ListEmails` filter logic |
| `cmd/list.go` | Add `--flagged`/`--unflagged` flags, validation, pass to `ListEmails` |
| `cmd/search.go` | Add `--flagged`/`--unflagged` flags, validation, set on `SearchOptions` |
| `internal/client/email_test.go` | Tests for filter construction with flagged/unflagged/compound cases |
| `cmd/list_test.go` | Mutual exclusion validation test |
| `docs/CLI-REFERENCE.md` | Document new flags |

## Verification

1. `go build ./...` compiles without errors
2. `go test ./...` passes all existing and new tests
3. `go vet ./...` clean
4. `jm list --flagged` / `jm search --flagged` return only flagged emails
5. `jm list --unflagged` / `jm search --unflagged` return only unflagged emails
6. `jm list --flagged --unflagged` returns a clear error
7. `jm list --unread --unflagged` works (compound filter)
8. `jm list --help` / `jm search --help` show the new flags
