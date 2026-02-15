# Plan: Fix Text Search and UX Improvements

Findings from a hands-on test of every `jm` command against a live Fastmail account on 2026-02-10. One bug prevents the most important search capability from working, and three UX gaps reduce the tool's effectiveness for CLI and LLM callers.

---

## Summary

| Category | Count | Items |
|----------|-------|-------|
| Bug | 1 | Full-text search broken |
| UX improvements | 2 | Accept bare dates in search, improve opaque error messages |
| Functional improvement | 1 | Add mark-as-read command |

---

## Phase 1: Bug Fix (Must Fix)

### 1.1 Fix full-text search (`jm search <query>`)

**Problem:** `jm search Burlington` returns `{"error":"jmap_error","message":"search: unknown returned invalidArguments"}`. All filter-only searches work fine (`--from`, `--subject`, `--has-attachment`, etc.). The failure only occurs when a positional text query is provided, which triggers the `SearchSnippet/get` code path.

**Observed behavior:**

```
$ jm search Burlington --limit 3
{"error":"jmap_error","message":"search: unknown returned invalidArguments"}

$ jm search --subject "Burlington" --limit 3
# Works fine, returns results
```

**Root cause analysis:** The previous plan (2026-02-08) identified and fixed an incorrect result reference in the `SearchSnippet/get` call -- it now correctly references `queryCallID` / `Email/query` / `/ids`. Despite that fix, the call still returns `invalidArguments` from the Fastmail JMAP server. The issue is likely in how the `filter` parameter is being passed to `SearchSnippet/get`, or in a field serialization issue in the `go-jmap` library's `searchsnippet.Get` type.

**Investigation steps:**

1. Enable HTTP-level request/response logging to see the exact JSON being sent to the JMAP server for the `SearchSnippet/get` call (redact `Authorization` headers and any bearer tokens before saving logs)
2. Compare the outgoing request against the JMAP spec (RFC 8621, Section 5.1) for `SearchSnippet/get` -- in particular, verify whether `filter` is being serialized with the correct property names
3. Check the `go-jmap` library's `searchsnippet.Get` struct to see if there's a field mapping or serialization issue (e.g., `Filter` field might need a different type than `email.FilterCondition`)
4. Test removing the `filter` parameter from the `SearchSnippet/get` call to see if the server accepts it without a filter (this would confirm the filter is the problem)
5. Test with a minimal filter (only `Text` set, no other fields) to rule out field combination issues

**File:** `internal/client/email.go`, lines 300-311

**Current code:**

```go
if hasTextSearch {
    req.Invoke(&searchsnippet.Get{
        Account: c.accountID,
        Filter:  filter,
        ReferenceIDs: &jmap.ResultReference{
            ResultOf: queryCallID,
            Name:     "Email/query",
            Path:     "/ids",
        },
    })
}
```

**Possible fixes (in order of likelihood):**

- **A)** The `searchsnippet.Get` `Filter` field may need a dedicated filter type or different serialization. Check if the `go-jmap` library defines a separate filter type for search snippets.
- **B)** The filter passed to `SearchSnippet/get` may need to contain only the text-related fields (stripping out `InMailbox`, `NotKeyword`, etc.). Create a minimal filter with only `Text` set for the snippet call.
- **C)** There may be a bug in the `go-jmap` library's JSON serialization for `searchsnippet.Get`. If so, construct the JMAP method call manually or patch the library.

**Verification:**

- `jm search Burlington --limit 3` should return results (currently fails)
- `jm search "meeting" --limit 3` should return results with `snippet` fields containing `<mark>` highlight tags
- `jm search Burlington --subject ice --limit 3` should work (combined text + filter)
- `jm search --from alice --limit 3` should continue working (no text query, no snippets)
- If `SearchSnippet/get` fails but `Email/query` + `Email/get` succeed, return results without snippets instead of failing the entire command

**Tests:**

- Add/extend unit tests in `internal/client/email_test.go` to verify:
  - Text search request wiring for `SearchSnippet/get` remains correct
  - Method error messages identify the failing method by call ID
  - Optional resilience behavior: snippet failure does not fail the whole search when query/get succeeded

---

## Phase 2: UX Improvements

### 2.1 Accept bare dates in `--before` and `--after` flags

**Problem:** The `--before` and `--after` flags on the `search` command require full RFC 3339 format (`2026-02-01T00:00:00Z`). Bare dates like `2026-02-01` are rejected with a parse error. This is especially inconvenient for LLM callers, which will naturally try the shorter form first and waste a round trip on the error.

**Current behavior:**

```
$ jm search --has-attachment --after 2026-02-01 --limit 3
{"error":"general_error","message":"invalid --after date: parsing time \"2026-02-01\" as \"2006-01-02T15:04:05Z07:00\": cannot parse \"\" as \"T\"","hint":"Use RFC 3339 format, e.g. 2026-01-15T00:00:00Z"}
```

**File:** `cmd/search.go`, lines 50-66

**Changes:**

1. Attempt `time.Parse(time.RFC3339, ...)` first, then fall back to parsing a bare date with `time.Parse("2006-01-02", ...)`
2. If the bare date parse succeeds, treat it as midnight UTC on that day
3. Keep the RFC 3339 parse as the primary path; the bare date is a fallback
4. Update the hint text to mention that bare dates are also accepted

**Implementation:**

```go
func parseDate(s string) (time.Time, error) {
    // Try RFC 3339 first (full precision).
    t, err := time.Parse(time.RFC3339, s)
    if err == nil {
        return t, nil
    }
    // Fall back to bare date (midnight UTC).
    t, err2 := time.Parse("2006-01-02", s)
    if err2 == nil {
        return t, nil
    }
    // Return the original RFC 3339 error for a better message.
    return time.Time{}, err
}
```

Apply this to both the `--before` and `--after` flag parsing blocks.

**Verification:**

- `jm search --after 2026-02-01 --limit 3` should work (bare date)
- `jm search --after 2026-02-01T00:00:00Z --limit 3` should continue working (RFC 3339)
- `jm search --before 2026-02-10 --after 2026-02-01 --limit 3` should work (both bare)
- `jm search --after "not-a-date" --limit 3` should still produce a clear error

**Tests:**

- Add CLI flag parsing tests to `tests/flags.md`:
  - `--before 2026-02-01` accepted
  - `--after 2026-02-01` accepted
  - Invalid dates still return a `general_error` with updated hint text

---

### 2.2 Improve error messages for JMAP method errors in search

**Problem:** When the `SearchSnippet/get` call fails, the error message is `"search: unknown returned invalidArguments"`. The `unknown` comes from `inv.Name` being empty or "error" in the response. This makes it unclear which of the three batched JMAP calls (Email/query, Email/get, SearchSnippet/get) failed.

**File:** `internal/client/email.go`, `SearchEmails` method, lines 322-341

**Current code:**

```go
case *jmap.MethodError:
    method := inv.Name
    if method == "" || method == "error" {
        method = "unknown"
    }
    return types.EmailListResult{}, fmt.Errorf("search: %s returned %s", method, r.Error())
```

**Changes:**

1. Track call IDs returned by `req.Invoke()` for each of the three method calls (query, get, snippet)
2. When a `MethodError` is encountered, match the response's call ID against the tracked IDs to determine which method failed
3. Include the method error's description if available (not just the error type)
4. Fall back to the invocation index if call ID matching isn't possible with the `go-jmap` response API

**Example -- before:**

```
"search: unknown returned invalidArguments"
```

**Example -- after:**

```
"search: SearchSnippet/get (call 2) returned invalidArguments"
```

Or, if the library exposes a description:

```
"search: SearchSnippet/get returned invalidArguments: The filter property 'foo' is not supported"
```

**Verification:** Trigger a method error and confirm the message identifies the failing method.

**Tests:**

- Add/extend unit tests in `internal/client/email_test.go` for `MethodError` reporting to ensure failures map to `Email/query`, `Email/get`, or `SearchSnippet/get` by call ID when invocation names are empty/opaque

---

## Phase 3: Functional Improvement

### 3.1 Add a `mark-read` command

**Problem:** There is no way to mark an email as read/seen without moving it. After reading an email with `jm read`, it remains marked as unread. For email triage workflows, marking as read is a fundamental operation alongside archive and spam.

**Safety considerations:** This is a non-destructive metadata update (setting the `$seen` keyword). It fits the tool's read-oriented philosophy -- it doesn't send, delete, or move anything.

**Files:**

- `cmd/mark-read.go` (new)
- `internal/client/email.go` (new method)

**Command design:**

```
jm mark-read <id> [id...]
```

- Accepts one or more email IDs (same pattern as `archive`, `spam`, `move`)
- Sets the `$seen` keyword on each email via `Email/set` update
- Uses the same batch processing pattern as `MoveEmails` (batch size 50)
- Returns structured JSON output matching the pattern of `archive`/`spam`/`move`

**Client method:**

```go
func (c *Client) MarkAsRead(emailIDs []string) (succeeded []string, errors []string) {
    // Same batch loop as MoveEmails
    updates[jmap.ID(id)] = jmap.Patch{
        "keywords/$seen": true,
    }
    // ... same error handling pattern
}
```

**Output format (JSON):**

```json
{
  "marked_as_read": ["id1", "id2"],
  "errors": []
}
```

**Output format (text):**

```
Marked as read: id1, id2
```

**Verification:**

- `jm mark-read <id>` should succeed and the email should no longer appear in `jm list --unread` results
- `jm mark-read <id1> <id2>` should handle multiple IDs
- `jm mark-read nonexistent-id` should return a structured error
- `jm mark-read` (no args) should fail with a usage error

**Tests:**

- Add argument validation test to `tests/arguments.md` (requires at least one ID)
- Add help output test to `tests/help.md`
- Add auth error test to `tests/errors.md`
- Add live test to `tests/live.md` (mark an email as read, verify with list --unread)

---

## Implementation Order

| Step | Item | Depends On |
|------|------|------------|
| 1 | 1.1 Fix full-text search | -- |
| 2 | 2.2 Improve JMAP error messages | Helpful context from 1.1 investigation |
| 3 | 2.1 Accept bare dates | -- |
| 4 | 3.1 Add mark-read command | -- |

Steps 1 and 3 can be done in parallel. Step 2 benefits from the investigation work in step 1. Step 4 is independent.

---

## Files Modified

| File | Changes |
|------|---------|
| `internal/client/email.go` | Fix SearchSnippet/get call (1.1), improve error messages (2.2), add MarkAsRead method (3.1) |
| `internal/client/email_test.go` | Add/extend tests for search snippets, method error mapping, and optional snippet-failure fallback |
| `cmd/search.go` | Accept bare dates in --before/--after (2.1) |
| `cmd/mark-read.go` | New file: mark-read command (3.1) |
| `internal/types/types.go` | Add `marked_as_read` field to operation result type (3.1) |
| `internal/output/text.go` | Format mark-read output (3.1) |
| `tests/help.md` | Add mark-read help test (3.1) |
| `tests/arguments.md` | Add mark-read argument validation (3.1) |
| `tests/errors.md` | Add mark-read auth error test (3.1) |
| `tests/flags.md` | Add bare date acceptance tests and updated date hint expectations (2.1) |
| `tests/live.md` | Add mark-read live test (3.1) |
| `docs/CLI-REFERENCE.md` | Document mark-read command, updated date flag behavior (2.1, 3.1) |
