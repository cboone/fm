# Bulk Operations: Respect JMAP Server Limits

## Context

Large email triage operations (archive, move, spam, mark-read, flag, unflag) can
fail when the number of IDs exceeds the JMAP server's `maxObjectsInSet` limit.
The current batch size is hardcoded at 50, which may be too large for some
servers or unnecessarily conservative for others. Additionally, the five bulk
operation functions duplicate ~55 lines of identical batching logic each, and
the JSON output lacks machine-friendly summary fields for automation consumers.

**GH Issue**: #27

## Changes

### 1. Read batch size from JMAP session capabilities

Add a method to `Client` that reads `MaxObjectsInSet` from the `core.Core`
capability in the JMAP session. Fall back to 50 if the capability is unavailable
or the value is 0.

**File**: `internal/client/client.go`

```go
import "git.sr.ht/~rockorager/go-jmap/core"

func (c *Client) maxBatchSize() int {
    if cap, ok := c.jmap.Session.Capabilities[jmap.CoreURI]; ok {
        if core, ok := cap.(*core.Core); ok && core.MaxObjectsInSet > 0 {
            return int(core.MaxObjectsInSet)
        }
    }
    return defaultBatchSize
}
```

- Rename the existing `batchSize` constant to `defaultBatchSize`
- The go-jmap library already parses `core.Core` from the session
  (`core/core.go:10-25`); we just need to read it

### 2. Centralize batch execution logic

Extract a shared `batchSetEmails` helper that all five bulk operations delegate
to. This eliminates ~250 lines of duplicated batching/error-accumulation code.

**File**: `internal/client/email.go`

```go
// batchSetEmails executes Email/set in server-aware batches.
// patchFn builds the jmap.Patch for a single email ID.
func (c *Client) batchSetEmails(emailIDs []string, patchFn func(string) jmap.Patch) (succeeded, errors []string) {
    size := c.maxBatchSize()
    succeeded = []string{}
    errors = []string{}

    for start := 0; start < len(emailIDs); start += size {
        end := start + size
        if end > len(emailIDs) {
            end = len(emailIDs)
        }
        batch := emailIDs[start:end]

        updates := make(map[jmap.ID]jmap.Patch, len(batch))
        for _, id := range batch {
            updates[jmap.ID(id)] = patchFn(id)
        }

        req := &jmap.Request{}
        req.Invoke(&email.Set{
            Account: c.accountID,
            Update:  updates,
        })

        resp, err := c.Do(req)
        if err != nil {
            for _, id := range batch {
                errors = append(errors, fmt.Sprintf("%s: %v", id, err))
            }
            continue
        }

        for _, inv := range resp.Responses {
            switch r := inv.Args.(type) {
            case *email.SetResponse:
                for _, idStr := range batch {
                    jid := jmap.ID(idStr)
                    if _, ok := r.Updated[jid]; ok {
                        succeeded = append(succeeded, idStr)
                    } else if setErr, ok := r.NotUpdated[jid]; ok {
                        desc := "unknown error"
                        if setErr.Description != nil {
                            desc = *setErr.Description
                        }
                        errors = append(errors, fmt.Sprintf("%s: %s", idStr, desc))
                    }
                }
            case *jmap.MethodError:
                for _, id := range batch {
                    errors = append(errors, fmt.Sprintf("%s: %s", id, r.Error()))
                }
            }
        }
    }
    return succeeded, errors
}
```

Each operation becomes a one-liner delegation:

```go
func (c *Client) MoveEmails(emailIDs []string, targetMailboxID jmap.ID) ([]string, []string) {
    return c.batchSetEmails(emailIDs, func(_ string) jmap.Patch {
        return jmap.Patch{"mailboxIds": map[jmap.ID]bool{targetMailboxID: true}}
    })
}

func (c *Client) MarkAsSpam(emailIDs []string, junkMailboxID jmap.ID) ([]string, []string) {
    return c.batchSetEmails(emailIDs, func(_ string) jmap.Patch {
        return jmap.Patch{
            "mailboxIds":     map[jmap.ID]bool{junkMailboxID: true},
            "keywords/$junk": true,
        }
    })
}

func (c *Client) MarkAsRead(emailIDs []string) ([]string, []string) {
    return c.batchSetEmails(emailIDs, func(_ string) jmap.Patch {
        return jmap.Patch{"keywords/$seen": true}
    })
}

func (c *Client) SetFlagged(emailIDs []string) ([]string, []string) {
    return c.batchSetEmails(emailIDs, func(_ string) jmap.Patch {
        return jmap.Patch{"keywords/$flagged": true}
    })
}

func (c *Client) SetUnflagged(emailIDs []string) ([]string, []string) {
    return c.batchSetEmails(emailIDs, func(_ string) jmap.Patch {
        return jmap.Patch{"keywords/$flagged": nil}
    })
}
```

### 3. Add summary fields to `MoveResult`

**File**: `internal/types/types.go`

Add three fields to `MoveResult`:

```go
type MoveResult struct {
    Matched   int    `json:"matched"`
    Processed int    `json:"processed"`
    Failed    int    `json:"failed"`
    // ... existing fields unchanged ...
}
```

These are computed in each command handler from the values already available:

```go
result := types.MoveResult{
    Matched:  len(args),
    Processed: len(succeeded),
    Failed:   len(errors),
    Archived: succeeded,
    Errors:   errors,
    // ...
}
```

**Files to update** (all command handlers that construct `MoveResult`):
- `cmd/archive.go`
- `cmd/move.go`
- `cmd/spam.go`
- `cmd/mark-read.go`
- `cmd/flag.go`
- `cmd/unflag.go`

### 4. Update text formatter for summary fields

**File**: `internal/output/text.go`

Add a summary line to `formatMoveResult`:

```go
fmt.Fprintf(w, "Matched: %d, Processed: %d, Failed: %d\n", r.Matched, r.Processed, r.Failed)
```

JSON formatter needs no changes (fields are auto-serialized).

### 5. Tests

**File**: `internal/client/email_test.go`

Add tests using the existing `doFunc` mock pattern (already used in
`TestSetFlagged_MixedUpdatedAndNotUpdated` etc.):

- **`TestBatchSetEmails_UsesServerMaxObjectsInSet`**: Create a client with a
  mock session that has `MaxObjectsInSet = 2`, send 5 IDs, verify 3 batches of
  sizes 2, 2, 1 are sent.
- **`TestBatchSetEmails_FallsBackToDefault`**: Create a client with no core
  capability, verify the default batch size is used.
- **`TestBatchSetEmails_LargeBatch`**: Send 200+ IDs, verify all are processed
  across batches with correct success/error accumulation.
- **`TestBatchSetEmails_PartialBatchFailure`**: First batch succeeds, second
  batch returns a network error, third batch succeeds. Verify succeeded and
  errors are correctly interleaved.
- **`TestBatchSetEmails_MixedPerIDErrors`**: A batch where some IDs succeed and
  some return `NotUpdated` errors. (Existing tests cover this for `SetFlagged`
  and `SetUnflagged`; extend to the shared helper.)
- **`TestMaxBatchSize_FromSession`**: Unit test for `maxBatchSize()` with
  various capability states.

**File**: `internal/client/client_test.go`

- **`TestMaxBatchSize_*`**: Tests for the `maxBatchSize()` helper.

Update existing tests that assert on `batchSize` constant
(`TestBatchSizeConstant`) to use `defaultBatchSize`.

### Files Modified (summary)

| File | Change |
|------|--------|
| `internal/client/client.go` | Add `maxBatchSize()`, import `core`, rename constant |
| `internal/client/email.go` | Add `batchSetEmails()`, simplify 5 bulk ops |
| `internal/types/types.go` | Add `Matched`, `Processed`, `Failed` to `MoveResult` |
| `internal/output/text.go` | Display summary line in `formatMoveResult` |
| `cmd/archive.go` | Populate summary fields |
| `cmd/move.go` | Populate summary fields |
| `cmd/spam.go` | Populate summary fields |
| `cmd/mark-read.go` | Populate summary fields |
| `cmd/flag.go` | Populate summary fields |
| `cmd/unflag.go` | Populate summary fields |
| `internal/client/email_test.go` | Add batch tests, update constant test |
| `internal/client/client_test.go` | Add `maxBatchSize` tests |

## Verification

1. **Unit tests**: `go test ./...` -- all new and existing tests pass
2. **Build**: `make build` succeeds
3. **CLI integration tests**: `make test-cli` passes (help text expectations unchanged)
4. **Manual smoke test** (if `JMAP_TOKEN` available): run `jm archive <id1> <id2>` and verify JSON output includes `matched`, `processed`, `failed` fields
