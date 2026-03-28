# Optimize resolveEmailIDs for single-result commands

Issue: #80

## Context

When single-email commands (like `unsubscribe`) use filter flags, `resolveEmailIDs` calls `QueryEmailIDs`, which paginates through the **entire** result set in batches of 250. Since only the first match is needed, this is unnecessarily expensive for broad filters. Adding a `QueryFirstEmailID` client method (with `Limit: 1`) and a corresponding `resolveFirstEmailID` command helper provides an efficient path for single-email commands.

## Classification

**Type:** refactor
**Commit prefix:** `refactor`

## Changes

### 1. Add `QueryFirstEmailID` to `internal/client/email.go`

Insert after `QueryEmailIDs` (after line 567), before the `StatsOptions` type.

```go
// QueryFirstEmailID runs Email/query with Limit 1 and returns the most recent
// matching email ID, sorted by receivedAt descending. If no emails match, it
// returns ("", nil).
func (c *Client) QueryFirstEmailID(opts SearchOptions) (string, error) {
```

Key design decisions:
- `Limit: 1` in the JMAP request, no pagination loop
- Explicit `Sort` by `receivedAt` descending (matches `SearchEmails`, `AggregateEmailsBySender`, `AggregateSummary` convention)
- No `CalculateTotal: true` since the total count isn't needed
- Returns `("", nil)` for no matches (caller decides the error message)
- Error format uses `"email/query: ..."` prefix, matching `QueryEmailIDs`

### 2. Add `resolveFirstEmailID` to `cmd/filters.go`

Insert after `resolveEmailIDs` (after line 175), before `parseDate`.

```go
// resolveFirstEmailID returns a single email ID from args or queries the most
// recent match using filter flags.
func resolveFirstEmailID(cmd *cobra.Command, args []string, c *client.Client) (string, error) {
```

Mirrors `resolveEmailIDs` in structure:
- Args check: returns `args[0]` immediately if args provided
- Calls `parseFilterOptions` then `c.QueryFirstEmailID(opts)`
- Same error codes: `"jmap_error"` and `"not_found"`

### 3. Add tests for `QueryFirstEmailID` in `internal/client/email_test.go`

Insert after `TestQueryEmailIDs_IgnoresLimitAndSort` (after line 2649), before `// --- extractDomain tests ---`.

Five tests following existing `TestQueryEmailIDs_*` patterns:

1. `TestQueryFirstEmailID_ReturnsFirst` - happy path, returns single ID
2. `TestQueryFirstEmailID_EmptyResult` - no matches returns `("", nil)`
3. `TestQueryFirstEmailID_MethodError` - JMAP error propagation
4. `TestQueryFirstEmailID_UsesLimitOneAndSort` - validates `Limit: 1`, sort by `receivedAt` desc, no `CalculateTotal`
5. `TestQueryFirstEmailID_UsesFilter` - validates filter passthrough

### 4. Add test for `resolveFirstEmailID` in `cmd/filters_test.go`

Insert after `TestParseFilterOptions_RecipientToFilterStillWorks` (after line 102).

One test: `TestResolveFirstEmailID_ReturnsFirstArg` - verifies args-passthrough returns `args[0]` with nil client (short-circuits before client use).

## Not in scope

- No existing command callers change. All 6 bulk commands (archive, spam, unflag, flag, mark-read, move) correctly use `resolveEmailIDs`/`QueryEmailIDs`.
- The `unsubscribe` command (first consumer of `resolveFirstEmailID`) is separate work on another branch.

## Verification

1. `go build ./...` compiles without errors
2. `go test ./...` passes all existing and new tests
3. `go vet ./...` reports no issues
