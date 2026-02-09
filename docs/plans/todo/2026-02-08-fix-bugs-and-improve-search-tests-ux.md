# Plan: Fix Bugs, Harden Tests, and Improve Search/UX

Findings from a full manual test of every `jm` command against a live Fastmail account on 2026-02-08. Two bugs prevent core functionality from working; the CLI integration test suite fails in any environment where `JMAP_TOKEN` is set; and several UX gaps limit the usefulness of the `search` command.

---

## Summary

| Category | Count | Items |
|----------|-------|-------|
| Bugs | 2 | Text search broken, scrut tests fail with token set |
| Functional improvements | 4 | Search pagination/sort/unread, mailbox caching |
| UX improvements | 2 | Sort syntax flexibility, opaque search error messages |
| Test improvements | 3 | Environment isolation, success-path tests, text format tests |

---

## Phase 1: Bugs (Must Fix)

### 1.1 Fix text search (`jm search "query"`)

**Problem:** `jm search "meeting"` returns `{"error":"jmap_error","message":"search: invalidArguments"}`. Filter-only search (e.g. `jm search --from alice`) works fine.

**Root cause:** When a text query is provided, `SearchEmails` in `internal/client/email.go` adds a `SearchSnippet/get` call with an incorrect JMAP result reference. It references the `Email/get` response at path `/list/*/id`, but `SearchSnippet/get` expects a flat array of email IDs from `Email/query` at path `/ids`.

**File:** `internal/client/email.go`, lines 294-302

**Current code:**

```go
req.Invoke(&searchsnippet.Get{
    Account: c.accountID,
    Filter:  filter,
    ReferenceIDs: &jmap.ResultReference{
        ResultOf: getCallID,
        Name:     "Email/get",
        Path:     "/list/*/id",
    },
})
```

**Fix:** Change the result reference to point at `Email/query`:

```go
req.Invoke(&searchsnippet.Get{
    Account: c.accountID,
    Filter:  filter,
    ReferenceIDs: &jmap.ResultReference{
        ResultOf: queryCallID,
        Name:     "Email/query",
        Path:     "/ids",
    },
})
```

**Verification:**
- `jm search "meeting" --limit 3` should return results with `snippet` fields containing `<mark>` tags
- `jm search "meeting" --from notifications@github.com --limit 3` should also work (combined text + filter)
- `jm search --from alice --limit 3` should continue working (no text query, no snippets)

---

### 1.2 Fix CLI integration tests (scrut) environment isolation

**Problem:** 26 of 47 scrut tests fail when `JMAP_TOKEN` is set in the environment. The tests expect `authentication_failed` errors but instead get successful API responses because `viper.AutomaticEnv()` picks up the ambient `JMAP_TOKEN`.

**Root cause:** The scrut test commands run as `$TESTDIR/../jm session 2>&1` without clearing environment variables. When `JMAP_TOKEN` is set in the shell, every test that expects "no token" errors instead succeeds.

**Files:** `tests/errors.md`, `tests/flags.md`, `tests/arguments.md`

**Fix:** For every test that expects an `authentication_failed` error (i.e. tests that depend on no token being present), prefix the command with `env -u JMAP_TOKEN -u JMAP_SESSION_URL -u JMAP_FORMAT -u JMAP_ACCOUNT_ID` to ensure a clean environment. Tests that explicitly set their own env vars (e.g. `JMAP_TOKEN=test-token`) should also clear the ambient token first.

**Example -- before:**

```scrut
$ $TESTDIR/../jm session 2>&1
{
  "error": "authentication_failed",
  ...
```

**Example -- after:**

```scrut
$ env -u JMAP_TOKEN -u JMAP_SESSION_URL -u JMAP_FORMAT -u JMAP_ACCOUNT_ID $TESTDIR/../jm session 2>&1
{
  "error": "authentication_failed",
  ...
```

For tests that set specific env vars (like `JMAP_TOKEN=test-token`), also clear other vars:

```scrut
$ env -u JMAP_SESSION_URL -u JMAP_FORMAT -u JMAP_ACCOUNT_ID JMAP_TOKEN=test-token $TESTDIR/../jm session 2>&1
```

**Specific tests to update:**

In `tests/errors.md` (9 tests):
- "Missing token produces structured JSON error" (line 8)
- "Missing token with text format" (line 20)
- "List without token" (line 29)
- "Read without token" (line 41)
- "Search without token" (line 53)
- "Archive without token" (line 65)
- "Spam without token" (line 76)
- "Move without token" (line 89)
- "Mailboxes without token" (line 101)

In `tests/flags.md` (15 tests):
- "Default format is JSON" (line 10)
- "Format flag switches to text" (line 22)
- "JMAP_FORMAT env var switches to text" (line 31)
- "JMAP_TOKEN env var is read" (line 43)
- "Token flag overrides env var" (line 53)
- "Custom session URL via flag" (line 63)
- "JMAP_SESSION_URL env var" (line 73)
- "List default flags" (line 85)
- "List with all flags" (line 95)
- "Search with all filter flags" (line 105)
- "Read with all flags" (line 138)
- "Mailboxes with roles-only flag" (line 148)
- "Archive with multiple IDs" (line 158)
- "Spam with multiple IDs" (line 168)
- "Move with multiple IDs" (line 178)

In `tests/arguments.md` (2 tests):
- "Search accepts zero arguments" (line 74)
- "Search accepts one argument" (line 85)

**Verification:** `make test-cli` should pass with 47/47 tests succeeding regardless of whether `JMAP_TOKEN` is set.

---

## Phase 2: Functional Improvements

### 2.1 Add `--offset` flag to `search` command

**Problem:** `list` supports `--offset` for pagination but `search` does not, making it impossible to paginate search results beyond the first page.

**File:** `cmd/search.go`

**Changes:**
1. Add `--offset` / `-o` flag (type `int64`, default `0`) to `searchCmd` in `init()`
2. Add `Offset` field to `SearchOptions` struct in `internal/client/email.go`
3. Pass `Position: opts.Offset` to `email.Query` in `SearchEmails`
4. Set `result.Offset = opts.Offset` in the response construction
5. Add input validation: offset must be non-negative (matching `list` behavior)

**Verification:** `jm search --subject "order" --limit 3 --offset 3` should return the next 3 results.

---

### 2.2 Add `--sort` flag to `search` command

**Problem:** Search hardcodes `receivedAt desc` sort order. Users can't sort search results by other fields.

**Files:** `cmd/search.go`, `internal/client/email.go`

**Changes:**
1. Add `--sort` / `-s` flag (type `string`, default `"receivedAt desc"`) to `searchCmd` in `init()`
2. Add `SortField` and `SortAsc` fields to `SearchOptions`
3. Parse sort string using the existing `parseSort()` function from `cmd/list.go` (move it to a shared location or keep it in `list.go` and call from both)
4. Pass sort values to `email.Query` in `SearchEmails`

**Note:** The `parseSort` function in `cmd/list.go` is currently unexported but package-private, so `search.go` in the same package can already call it directly.

**Verification:** `jm search --subject "order" --sort "subject asc" --limit 3` should return results sorted by subject ascending.

---

### 2.3 Add `--unread` flag to `search` command

**Problem:** `list` supports `--unread` to filter to unread-only, but `search` does not.

**Files:** `cmd/search.go`, `internal/client/email.go`

**Changes:**
1. Add `--unread` / `-u` flag (type `bool`, default `false`) to `searchCmd` in `init()`
2. Add `UnreadOnly` field to `SearchOptions`
3. In `SearchEmails`, when `opts.UnreadOnly` is true, set `filter.NotKeyword = "$seen"` (same approach as `ListEmails`)

**Verification:** `jm search --from notifications@github.com --unread --limit 3` should return only unread results.

---

### 2.4 Cache mailbox list within a single command invocation

**Problem:** `ResolveMailboxID` can make two separate `Mailbox/get` API calls: first via `GetMailboxByRole`, then via `GetMailboxByNameOrID` if the role lookup fails. Commands like `archive`, `spam`, and `move` also call mailbox resolution for safety validation. Each call fetches all mailboxes from the server.

**File:** `internal/client/mailbox.go`

**Changes:**
1. Add a `mailboxCache []*mailbox.Mailbox` field to the `Client` struct
2. In `GetAllMailboxes`, check the cache first; if populated, return cached data
3. After a successful `Mailbox/get` response, store results in the cache
4. All existing callers (`GetMailboxByRole`, `GetMailboxByNameOrID`, `ListMailboxes`) already go through `GetAllMailboxes`, so they automatically benefit

This is a per-`Client`-instance cache (no disk persistence), so it only saves redundant calls within a single command invocation.

**Verification:** Existing unit tests should still pass. Functional behavior is unchanged; the improvement is fewer API round trips.

---

## Phase 3: UX Improvements

### 3.1 Accept colon syntax in `--sort` flag

**Problem:** Users may try `--sort "receivedAt:asc"` (colon-separated) but only space-separated `"receivedAt asc"` is accepted. The error message says `unsupported sort field "receivedAt:asc"` which is confusing.

**File:** `cmd/list.go`, function `parseSort`

**Changes:**
1. Before splitting on whitespace, replace `:` with a space in the sort string
2. This makes `"receivedAt:asc"`, `"receivedAt:desc"`, and `"receivedAt asc"` all work
3. Add/adjust unit tests in `cmd/list_test.go` to cover colon syntax explicitly

```go
func parseSort(s string) (field string, ascending bool, err error) {
    s = strings.ReplaceAll(s, ":", " ")
    parts := strings.Fields(s)
    // ... rest unchanged
}
```

**Verification:**
- `jm list -m inbox --limit 3 -s "receivedAt:asc"` should work
- `jm list -m inbox --limit 3 -s "receivedAt asc"` should continue working
- `jm list -m inbox --limit 3 -s "receivedAt desc"` should continue working

---

### 3.2 Improve error messages for JMAP method errors in search

**Problem:** When search fails, the error message is `"search: invalidArguments"` with no hint. After fixing the text search bug (1.1), this specific error should no longer occur, but other JMAP errors could still surface with equally opaque messages.

**File:** `internal/client/email.go`, `SearchEmails` method

**Changes:**
1. When a `*jmap.MethodError` is encountered, include the method name in the error message to help identify which JMAP call failed
2. Check if the `MethodError` has a description field and include it

**Example -- before:**

```
"search: invalidArguments"
```

**Example -- after:**

```
"search: SearchSnippet/get returned invalidArguments: <description if available>"
```

**Implementation detail:** The response loop processes multiple method responses. Track which call ID corresponds to which method name, or use the invocation name from the response to identify the source.

---

## Phase 4: Test Improvements

### 4.1 Add success-path scrut tests

**Problem:** All existing scrut tests verify error conditions. There are no integration tests for successful operations.

**File:** Create `tests/live.md` (or similar name indicating these require a token)

**Approach:** Add a new test file for opt-in live integration tests. Keep default scrut runs deterministic.

- `make test-cli` should continue running only deterministic tests (`tests/errors.md`, `tests/flags.md`, `tests/arguments.md`, `tests/help.md`)
- Add a separate target (for example `make test-cli-live`) that runs `tests/live.md`
- Gate `tests/live.md` behind both an explicit opt-in env var (for example `JMAP_LIVE_TESTS=1`) and `JMAP_TOKEN` presence

Use a guard at the top:

```scrut
$ test "$JMAP_LIVE_TESTS" = "1" || exit 80
$ test -n "$JMAP_TOKEN" || exit 80
```

(Exit code 80 causes scrut to skip the test when live test preconditions are not met.)

**Tests to add:**
- `jm session` returns JSON with `username` field
- `jm session --format text` returns text starting with `Username:`
- `jm mailboxes` returns JSON array with at least one mailbox
- `jm mailboxes --roles-only` returns JSON array where every entry has a `role` field
- `jm list --limit 1` returns JSON with `total` and `emails` fields
- `jm list --limit 1 --format text` returns text starting with `Total:`
- `jm search --limit 1 --from <known sender>` returns results (filter-only)
- `jm search "the" --limit 1` returns results with `snippet` field (after bug 1.1 is fixed)
- `jm read <known email id>` returns JSON with `body` field (may need to discover an ID first via list)

**Note:** These tests depend on account state and will need to use glob patterns for variable fields (IDs, counts, dates).

**Verification:** `make test-cli` remains stable in any environment. `make test-cli-live` runs when explicitly opted in and skips cleanly when preconditions are not met.

---

### 4.2 Add text format scrut tests

**Problem:** CLI integration tests don't exercise `--format text` for successful output.

**File:** Add to `tests/live.md` (same file as 4.1, or a separate `tests/text-format.md`)

**Tests to add:**
- `jm session --format text` output contains `Username:`
- `jm mailboxes --format text` output contains column-aligned mailbox names
- `jm list --limit 1 --format text` output contains `Total:` header and `ID:` lines
- Error output in text format: `Error [<code>]:` prefix

---

### 4.3 Add unit tests for search with text query

**Problem:** No unit test covers the `SearchSnippet/get` code path in `SearchEmails`.

**File:** `internal/client/email_test.go`

**Changes:** Add a test that constructs a `SearchOptions` with `Text` set and verifies:
1. The JMAP request contains three method invocations (Email/query, Email/get, SearchSnippet/get)
2. The `SearchSnippet/get` call references `Email/query` (not `Email/get`)
3. The reference path is `/ids` (not `/list/*/id`)

This requires either inspecting the constructed `jmap.Request` before sending, or mocking the `Do()` method. Follow the existing test patterns in `email_test.go`.

---

## Implementation Order

| Step | Phase | Item | Depends On |
|------|-------|------|------------|
| 1 | 1.1 | Fix text search result reference | -- |
| 2 | 4.3 | Add unit test for search snippet reference | 1.1 |
| 3 | 1.2 | Fix scrut test environment isolation | -- |
| 4 | 2.1 | Add `--offset` to search | -- |
| 5 | 2.2 | Add `--sort` to search | -- |
| 6 | 2.3 | Add `--unread` to search | -- |
| 7 | 3.1 | Accept colon syntax in sort | -- |
| 8 | 2.4 | Cache mailbox list | -- |
| 9 | 3.2 | Improve JMAP error messages | 1.1 |
| 10 | 4.1 | Add success-path scrut tests | 1.1, 1.2 |
| 11 | 4.2 | Add text format scrut tests | 4.1 |

Steps 1-3 can be done in parallel. Steps 4-8 can be done in parallel after step 1. Steps 10-11 should come last.

---

## Files Modified

| File | Changes |
|------|---------|
| `internal/client/email.go` | Fix SearchSnippet reference (1.1), add offset/sort/unread to SearchOptions (2.1-2.3), improve error messages (3.2) |
| `cmd/search.go` | Add --offset, --sort, --unread flags (2.1-2.3) |
| `cmd/list.go` | Accept colon syntax in parseSort (3.1) |
| `internal/client/mailbox.go` | Add per-instance mailbox cache (2.4) |
| `internal/client/client.go` | Add cache field to Client struct (2.4) |
| `tests/errors.md` | Add environment isolation (1.2) |
| `tests/flags.md` | Add environment isolation (1.2) |
| `tests/arguments.md` | Add environment isolation (1.2) |
| `tests/help.md` | Update search help snapshot for new search flags (2.1-2.3) |
| `tests/live.md` | New file: success-path and text format tests (4.1, 4.2) |
| `internal/client/email_test.go` | Add search snippet reference test (4.3) |
| `cmd/list_test.go` | Add parseSort colon syntax coverage (3.1) |
| `docs/CLI-REFERENCE.md` | Update search command docs to reflect new flags (2.1-2.3) |
| `Makefile` | Add opt-in live scrut target (4.1) |

---

## What This Plan Does NOT Cover

These items were identified during testing but are out of scope for this plan:

- **Mailbox hierarchy visualization** -- The `parent_id` field exists but nested display is not implemented. This is a feature addition, not a bug fix.
- **`--quiet` flag for scripting** -- Nice to have but not blocking any workflow.
- **Attachment download** -- Explicitly out of scope per the original design.
- **Session caching** -- The original plan explicitly deferred this; the per-invocation mailbox cache (2.4) addresses the most common redundant-call pattern.
