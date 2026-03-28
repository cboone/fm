# Add `--from` flag to `fm draft`

## Context

`fm draft` currently hardcodes the From address from `Session.Username` (the Fastmail primary email). Users with multiple Fastmail identities/aliases (e.g., `cboone@fea.st`, `cboone@sent.com`, `chris@hypsography.com`) have no way to specify which From address to use when creating drafts. This matters for use cases like batch-creating unsubscribe drafts, where the From address must match the address subscribed to the list.

Issue: #69

## Approach

Add a `--from` flag that resolves against JMAP identities via `Identity/get`. Follow the existing mailbox resolution pattern (fetch-all, cache, resolve by email). When `--from` is omitted, fall back to current behavior (session username).

## Changes

### 1. New file: `internal/client/identity.go`

Follow the `mailbox.go` pattern:

- `GetAllIdentities() ([]*identity.Identity, error)`: fetch via `identity.Get{Account: c.accountID}`, cache in `c.identityCache`, return cached on subsequent calls
- `ResolveIdentityByEmail(email string) (*identity.Identity, error)`: call `GetAllIdentities()`, case-insensitive email match. On no match, return error listing available identity emails

Import `git.sr.ht/~rockorager/go-jmap/mail/identity`. This transitively registers `emailsubmission` methods via `init()`. Add a comment explaining this is safe: Identity/get is read-only, and the codebase structurally never constructs `EmailSubmission/set` calls.

### 2. Modify `internal/client/client.go`

Add `identityCache []*identity.Identity` field to `Client` struct. Add the identity import.

### 3. Modify `internal/client/draft.go`

- Add `From string` field to `DraftOptions`
- In `CreateDraft`, replace the current From derivation block (lines 53-60) with:
  - If `opts.From != ""`: resolve via `c.ResolveIdentityByEmail(opts.From)`, set `fromAddrs` from resolved identity (including `Name`)
  - Else: keep current session username fallback
- The reply-all self-exclusion (line 96) already reads from `fromAddrs[0].Email`, so it works correctly with either path

### 4. Modify `cmd/draft.go`

- Register `--from` flag in `init()`: `draftCmd.Flags().String("from", "", "sender identity email address")`
- Read the flag value in `RunE` and pass as `From` in `DraftOptions`
- The error from `ResolveIdentityByEmail` routes through existing error handling in `RunE`

### 5. New file: `internal/client/identity_test.go`

Tests for:
- `GetAllIdentities` success, caching, error propagation
- `ResolveIdentityByEmail` exact match, case-insensitive match, no-match error with available list

### 6. Modify `internal/client/draft_test.go`

New tests:
- Draft with `From` resolves identity and uses it
- Draft with `From` no-match returns error
- Draft with empty `From` falls back to session username
- Reply-all with `From` excludes the resolved identity from CC

Update `testClientForDraft` to include an `identityCache` with test identities.

### 7. Update docs

- `docs/CLI-REFERENCE.md`: add `--from` row to draft flag table (between `--forward` and `--html`), add "From identity" section to description, add `--from` to example commands
- `tests/help.md`: add `*--from*` line to the draft help snapshot (between `--forward` and `--help`, maintaining alphabetical order)

## Key files

- `internal/client/mailbox.go` (reference pattern for identity.go)
- `internal/client/draft.go` (main logic changes)
- `cmd/draft.go` (flag registration)
- `internal/client/client.go` (add cache field)
- `internal/types/types.go` (no changes needed; From already in DraftResult)
- `internal/output/text.go` (no changes needed; already formats From)
- Go module: `git.sr.ht/~rockorager/go-jmap@v0.5.3/mail/identity/` (upstream types)

## Verification

1. `go build ./...` compiles
2. `go vet ./...` passes
3. `go test ./internal/client/...` passes (existing + new tests)
4. `go test ./...` passes (including scrut help tests with updated snapshots)
5. `make lint` passes
6. Manual test: `fm draft --from alias@example.com --to ... --subject ... --body ...` creates draft with correct From
7. Manual test: `fm draft --from nonexistent@example.com ...` returns error listing available identities
8. Manual test: `fm draft --to ... --subject ... --body ...` (no `--from`) still uses session username
