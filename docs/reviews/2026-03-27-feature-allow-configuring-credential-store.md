## Branch Review: feature/allow-configuring-credential-store

Base: main (merge base: 11f69f9)
Commits: 1
Files changed: 34 (1 added, 33 modified, 0 deleted, 0 renamed)
Reviewed through: 9703513

### Summary

This branch replaces the static `--token`/`FM_TOKEN` authentication mechanism with a configurable `--credential-command`/`FM_CREDENTIAL_COMMAND` that executes a user-specified shell command to retrieve the API token at runtime. Platform-specific defaults automatically retrieve the token from the OS keychain on macOS (Keychain via `security`) and Linux (libsecret via `secret-tool`), providing a zero-configuration experience on those platforms. This is a security improvement: tokens are no longer stored in environment variables or config files in plaintext.

### Changes by Area

**Core Authentication Logic (`cmd/root.go`)**

The central change. Adds `defaultCredentialCommand()` (platform-specific keychain defaults) and `resolveToken()` (executes the credential command via `sh -c`, captures stdout/stderr, trims whitespace). The `newClient()` function now calls `resolveToken()` instead of reading a static `token` viper key. The `--token` persistent flag is replaced with `--credential-command`, and the viper binding changes from `token` to `credential_command`.

- `cmd/root.go`

**Error Messages (all command files)**

Every subcommand's `authentication_failed` hint is updated from "Check your token in FM_TOKEN or config file" to "Check your credential command or the token it returns".

- `cmd/archive.go`, `cmd/draft.go`, `cmd/flag.go`, `cmd/list.go`, `cmd/mailboxes.go`, `cmd/mark-read.go`, `cmd/move.go`, `cmd/read.go`, `cmd/search.go`, `cmd/session.go`, `cmd/spam.go`, `cmd/stats.go`, `cmd/summary.go`, `cmd/unflag.go`
- `cmd/sieve_activate.go`, `cmd/sieve_create.go`, `cmd/sieve_deactivate.go`, `cmd/sieve_delete.go`, `cmd/sieve_list.go`, `cmd/sieve_show.go`, `cmd/sieve_validate.go`

**Tests**

All scrut test files are updated to replace `FM_TOKEN` with `FM_CREDENTIAL_COMMAND` and `--token` with `--credential-command`. Error message expectations are updated to use glob patterns for the new "credential command failed" messages. The Go unit test (`cmd/dryrun_test.go`) changes `--token test-token` to `--credential-command "echo test-token"`.

- `tests/errors.md`, `tests/flags.md`, `tests/arguments.md`, `tests/live.md`, `tests/sieve.md`
- `cmd/dryrun_test.go`

**Documentation**

README, Claude Code guide, CLI reference, and `.env.example` are all updated to document the new credential command approach, including platform-specific keychain storage instructions and examples of third-party secret store integration (1Password CLI, pass).

- `README.md`, `docs/CLAUDE-CODE-GUIDE.md`, `docs/CLI-REFERENCE.md`, `.env.example`

**Build**

Makefile comment updated to reference `FM_CREDENTIAL_COMMAND` instead of `FM_TOKEN`.

- `Makefile`

### File Inventory

- **New files**: 1 (docs/reviews/2026-03-27-feature-allow-configuring-credential-store.md)
- **Modified files**: 33 (.env.example, Makefile, README.md, cmd/archive.go, cmd/draft.go, cmd/dryrun_test.go, cmd/flag.go, cmd/list.go, cmd/mailboxes.go, cmd/mark-read.go, cmd/move.go, cmd/read.go, cmd/root.go, cmd/search.go, cmd/session.go, cmd/sieve_activate.go, cmd/sieve_create.go, cmd/sieve_deactivate.go, cmd/sieve_delete.go, cmd/sieve_list.go, cmd/sieve_show.go, cmd/sieve_validate.go, cmd/spam.go, cmd/stats.go, cmd/summary.go, cmd/unflag.go, docs/CLAUDE-CODE-GUIDE.md, docs/CLI-REFERENCE.md, tests/arguments.md, tests/errors.md, tests/flags.md, tests/live.md, tests/sieve.md)
- **Deleted files**: 0
- **Renamed files**: 0

### Notable Changes

- **Breaking change**: `--token` flag and `FM_TOKEN` environment variable are removed entirely. Users must migrate to `FM_CREDENTIAL_COMMAND` or rely on the platform default (macOS/Linux only). There is no deprecation period or backward compatibility shim.
- **Configuration change**: The config file key changes from `token` to `credential_command`. Existing `~/.config/fm/config.yaml` files with a `token` key will silently stop working.
- **Security improvement**: Tokens are no longer passed directly as environment variables or config values. The credential command pattern delegates secret storage to purpose-built tools.
- **New runtime dependency**: `sh` is invoked to execute credential commands. This is standard on all target platforms but worth noting.

### Code Quality Assessment

#### Code Quality

**Readability**: The implementation is clean and easy to follow. `defaultCredentialCommand()` and `resolveToken()` are well-named, concise, and have clear single responsibilities. The `switch` on `runtime.GOOS` is straightforward.

**Maintainability**: Good. Adding support for a new platform is a one-line addition to the `switch` statement. The credential command abstraction is a clean seam: `newClient()` doesn't know or care how the token is obtained.

**Patterns and consistency**: The new code follows the existing patterns in `cmd/root.go` perfectly. The error handling style (returning `fmt.Errorf` with context) matches the rest of the codebase. The hint string updates across all subcommands are consistent.

**Duplication**: The hint string "Check your credential command or the token it returns" is repeated in every subcommand file. This was already the case with the previous hint string, so it is not new duplication, but it is a pre-existing pattern worth noting.

#### Potential Issues

1. **Token executed on every command invocation**: `resolveToken()` shells out to the credential command on every CLI invocation. For commands run in tight loops, this could introduce latency (e.g., if the credential command hits a remote service like 1Password CLI). This is a reasonable tradeoff for a CLI tool, but worth being aware of.

2. **Fixed 10s timeout on credential command execution**: Credential commands run with a 10-second timeout via `context.WithTimeout`. This prevents indefinite hangs (e.g., waiting for a GUI prompt from a password manager). The fixed timeout may be too short for some setups; a future enhancement could make the timeout configurable.

3. **Shell injection surface**: The credential command is executed via `sh -c`, which means the value from `FM_CREDENTIAL_COMMAND`, `--credential-command`, or the config file is interpreted as shell code. This is intentional and documented (the user controls the input), but it is worth confirming that no untrusted source can set these values in practice. Since all three sources (env var, CLI flag, config file) require the user or their environment to set them, this is acceptable.

4. **Error message change may break test expectations on macOS/Linux defaults**: When no `FM_CREDENTIAL_COMMAND` is set but the platform is macOS or Linux, the default keychain command will be tried and likely fail with a "credential command failed" error (not the "no credential command configured" error). The scrut tests handle this correctly with glob patterns, but users on macOS/Linux who previously saw "no token configured" will now see a different error about the keychain command failing. This is a UX change, not a bug.

#### Completeness

- No TODO, FIXME, or placeholder code in the new implementation.
- Tests are comprehensively updated: error tests, flag tests, argument tests, live tests, sieve tests, and Go unit tests all reflect the new interface.
- Documentation is thoroughly updated across README, CLI reference, Claude Code guide, and `.env.example`.
- The config file example in README is updated, and the "Security note" about keeping tokens in env vars is correctly removed (since that advice no longer applies with credential commands).

#### Assessment Verdict

**Overall quality**: This code is ready to merge. The change is well-scoped, well-tested, and well-documented.

**Strengths**:
- Clean, minimal implementation that solves a real security concern.
- Platform defaults provide excellent out-of-the-box UX on macOS and Linux.
- Comprehensive update across all touchpoints: code, tests, docs, examples.
- Good error messages that distinguish between "no command configured" and "command failed".
- Proper stderr capture for debugging credential command failures.

**Issues to address**:
- None blocking. The implementation is solid.

**Suggestions** (non-blocking, resolved):
- ~~Consider adding a note in the README or a `--help` flag description about the platform defaults being used when no explicit credential command is set.~~ Resolved in 7364f06: flag help text now mentions OS keychain default.
- ~~A future enhancement could add a `--credential-command-timeout` or a hardcoded reasonable timeout to prevent indefinite hangs from misbehaving credential commands.~~ Resolved in 7364f06: added 10-second timeout via `context.WithTimeout`.
- Text-format scrut tests fixed in 13e1b57: added `glob*` patterns to accommodate optional stderr from credential commands.
