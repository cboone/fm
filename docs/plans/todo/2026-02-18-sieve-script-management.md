# Sieve Script Management (Issue #34)

## Context

There is no way to manage Fastmail sieve/filtering rules via `fm`. During spam triage, repeat offenders need to be blocked at the filter level, which currently requires the Fastmail web UI. This plan adds a `fm sieve` command group for managing sieve scripts via the JMAP SieveScript extension (RFC 9661).

**Key constraint**: The JMAP API operates on complete sieve scripts, not individual rules within a script. Only one script can be active per account at a time. The command group is named `sieve` to accurately reflect this model.

**Key constraint**: The go-jmap library (v0.5.3) has no sieve support. We must implement the JMAP SieveScript types as a local extension package following the library's existing patterns.

## Commands

| Command | Usage | Type |
|---------|-------|------|
| `sieve list` | `fm sieve list` | read-only |
| `sieve show` | `fm sieve show <script-id>` | read-only |
| `sieve validate` | `fm sieve validate [--script-stdin]` | read-only |
| `sieve create` | `fm sieve create --name "..." --from "..." --action junk` | mutation |
| `sieve create` | `fm sieve create --name "..." --script-stdin` | mutation |
| `sieve activate` | `fm sieve activate <script-id>` | mutation |
| `sieve deactivate` | `fm sieve deactivate` | mutation |
| `sieve delete` | `fm sieve delete <script-id>` | mutation |

All mutation commands support `--dry-run`.

`sieve create` supports two modes:
- **Template mode**: `--from` / `--from-domain` + `--action` generates a sieve script from a template
- **Stdin mode**: `--script-stdin` reads raw sieve from stdin

`sieve create --activate` optionally activates the script on creation. Without it, scripts are created inactive (safe default).

## New Files

### 1. JMAP Sieve Extension Types: `internal/jmap/sieve/`

Local extension package following the go-jmap pattern (`mail/vacationresponse/` as reference).

- **`sieve.go`**: URI constant (`urn:ietf:params:jmap:sieve`), `Sieve` capability struct, `SieveScript` data type, `init()` registering capability + methods
- **`get.go`**: `Get` request + `GetResponse` (List of `*SieveScript`, NotFound)
- **`set.go`**: `Set` request (Create/Update/Destroy + `OnSuccessActivateScript *jmap.ID` + `OnSuccessDeactivateScript *bool`) + `SetResponse`
- **`query.go`**: `Query` request + `QueryResponse`
- **`validate.go`**: `Validate` request (BlobID) + `ValidateResponse` (Error)

Each method type implements `Name() string` and `Requires() []jmap.URI`. The `init()` function calls `jmap.RegisterCapability` and `jmap.RegisterMethod` for each method.

### 2. Client Methods: `internal/client/sieve.go`

Blank import of `internal/jmap/sieve` to trigger `init()` registration.

Methods:
- `ListSieveScripts() (types.SieveScriptListResult, error)` -- SieveScript/get with nil IDs (returns all)
- `GetSieveScript(id string) (types.SieveScriptDetail, error)` -- SieveScript/get + blob download for content
- `GetSieveScriptContent(blobID string) (string, error)` -- blob download helper
- `CreateSieveScript(name, content string, activate bool) (types.SieveCreateResult, error)` -- blob upload + SieveScript/set create, optionally with `onSuccessActivateScript`
- `ValidateSieveScript(content string) (types.SieveValidateResult, error)` -- blob upload + SieveScript/validate
- `ActivateSieveScript(id string) error` -- SieveScript/set with `onSuccessActivateScript`
- `DeactivateSieveScript() error` -- SieveScript/set with `onSuccessDeactivateScript: true`
- `DeleteSieveScript(id string) (types.SieveDeleteResult, error)` -- SieveScript/set destroy (error if active)

**Blob handling**: Upload via `c.jmap.Upload(c.accountID, reader)`, download via `c.jmap.Download(c.accountID, blobID)`. Both methods exist on the underlying `jmap.Client`.

**Capability check**: Each method should verify `urn:ietf:params:jmap:sieve` is in `c.jmap.Session.RawCapabilities` and return a clear error if missing.

### 3. Template Generator: `internal/client/sieve_template.go`

Generates valid sieve scripts from flags:

```go
type SieveTemplateOptions struct {
    From       string // exact sender address match
    FromDomain string // sender domain match
    Action     string // junk | discard | keep | fileinto
    FileInto   string // mailbox name (when action=fileinto)
}

func GenerateSieveScript(opts SieveTemplateOptions) (string, error)
```

Example output for `--from "spam@example.com" --action junk`:
```sieve
require ["fileinto"];

if address :is "from" "spam@example.com" {
    fileinto "Junk";
    stop;
}
```

Example for `--from-domain "example.com" --action junk`:
```sieve
require ["fileinto"];

if address :domain :is "from" "example.com" {
    fileinto "Junk";
    stop;
}
```

### 4. Output Types: additions to `internal/types/types.go`

- `SieveScriptInfo` -- id, name, is_active (for list rows)
- `SieveScriptListResult` -- total + []SieveScriptInfo
- `SieveScriptDetail` -- id, name, blob_id, is_active, content
- `SieveCreateResult` -- id, name, blob_id, is_active, content
- `SieveDeleteResult` -- id, name
- `SieveActivateResult` -- id, name, is_active
- `SieveValidateResult` -- valid, error, content
- `SieveDryRunResult` -- operation, script name, content, valid

### 5. Text Formatting: additions to `internal/output/text.go`

Add type-switch cases for each new type. Key formats:
- **list**: tabwriter table with ID, Name, Active columns
- **show**: metadata header + separator + raw script content
- **create/delete/activate**: operation summary line

### 6. Command Files: `cmd/sieve*.go`

- **`sieve.go`**: parent command (no RunE, just help), `rootCmd.AddCommand(sieveCmd)`
- **`sieve_list.go`**: `sieveCmd.AddCommand(sieveListCmd)`
- **`sieve_show.go`**: takes script-id arg
- **`sieve_validate.go`**: `--script-stdin` or `--script` flag
- **`sieve_create.go`**: template flags (`--from`, `--from-domain`, `--action`, `--fileinto`, `--name`) or `--script-stdin`; `--activate` and `--dry-run`
- **`sieve_activate.go`**: takes script-id arg, `--dry-run`
- **`sieve_deactivate.go`**: no args, `--dry-run`
- **`sieve_delete.go`**: takes script-id arg, `--dry-run`

### 7. Tests

- **`internal/client/sieve_test.go`**: unit tests using mock `doFunc` for all client methods
- **`internal/client/sieve_template_test.go`**: template generation for each flag combination
- **`tests/sieve.md`**: scrut CLI integration tests (help output, auth errors, flag validation)

## Safety Measures

1. **No automatic activation**: `create` defaults to inactive; requires explicit `--activate`
2. **Cannot delete active scripts**: `delete` returns error if script is active; user must deactivate first
3. **Dry-run for all mutations**: create, activate, deactivate, delete
4. **Capability check**: clear error + hint if sieve capability not available on the server/token
5. **Server-side validation**: `create --dry-run` validates the script via SieveScript/validate

## Implementation Order

1. `internal/jmap/sieve/` package (types, registration)
2. `internal/client/sieve_template.go` + tests
3. Output types in `internal/types/types.go`
4. Client methods in `internal/client/sieve.go` + tests
5. Text formatting in `internal/output/text.go`
6. `cmd/sieve.go` parent command
7. Read-only commands: `sieve_list.go`, `sieve_show.go`, `sieve_validate.go`
8. Mutation commands: `sieve_create.go`, `sieve_activate.go`, `sieve_deactivate.go`, `sieve_delete.go`
9. Scrut tests in `tests/sieve.md`

## Key Files to Modify

- `internal/types/types.go` -- add sieve output types
- `internal/output/text.go` -- add sieve text formatting cases
- `internal/output/json.go` -- should work without changes (generic JSON marshal)

## Key Files for Reference

- `cmd/draft.go` -- complex mutation command pattern
- `cmd/move.go` -- mutation + dry-run pattern
- `internal/client/client.go` -- Client struct, Upload/Download access via `c.jmap`
- `internal/client/draft.go` -- blob-based creation pattern (email drafts)
- `internal/client/safety.go` -- safety validation pattern
- `/Users/hpg/go/pkg/mod/git.sr.ht/~rockorager/go-jmap@v0.5.3/mail/vacationresponse/` -- simple JMAP extension reference

## Verification

1. `make test` -- all unit tests pass
2. `make test-cli` -- scrut integration tests pass (help text, error cases)
3. `make vet && make fmt` -- code quality checks pass
4. Manual: `fm sieve list` against a live Fastmail account (requires `FM_TOKEN` with sieve scope)
5. Manual: `fm sieve create --name "test" --from "test@example.com" --action junk --dry-run` to verify template generation + server-side validation
