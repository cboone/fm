# Plan: Rename jm to fm (JMAP Mail to Fastmail Mail)

## Context

The CLI tool is being rebranded from "jm" (JMAP Mail) to "fm" (Fastmail Mail). The GitHub repository will also be renamed from `cboone/jm` to `cboone/fm`. This affects the binary name, Go module path, CLI command name, config directory, environment variable prefix (`JMAP_` to `FM_`), and all documentation and test files. Historical plan documents in `docs/plans/done/` are left as-is per project convention.

## Decisions

- **Binary name:** `jm` -> `fm`
- **Go module:** `github.com/cboone/jm` -> `github.com/cboone/fm`
- **CLI command name:** `jm` -> `fm`
- **Config directory:** `~/.config/jm/` -> `~/.config/fm/`
- **Env prefix:** `JMAP_` -> `FM_` (i.e., `FM_TOKEN`, `FM_SESSION_URL`, `FM_FORMAT`, `FM_ACCOUNT_ID`)
- **Descriptions:** Fastmail-branded (JMAP mentioned only as the protocol)
- **Old plans:** Leave unchanged (historical records)

## Implementation Steps

### Step 1: Go module path (`go.mod`)

**File:** `go.mod`

Change line 1:
- `module github.com/cboone/jm` -> `module github.com/cboone/fm`

Then run `go mod tidy` to regenerate `go.sum`.

### Step 2: All Go import paths (26+ files)

Every `.go` file that imports `github.com/cboone/jm/...` needs updating to `github.com/cboone/fm/...`. Use find-and-replace across all `.go` files:

- `"github.com/cboone/jm/cmd"` -> `"github.com/cboone/fm/cmd"`
- `"github.com/cboone/jm/internal/client"` -> `"github.com/cboone/fm/internal/client"`
- `"github.com/cboone/jm/internal/output"` -> `"github.com/cboone/fm/internal/output"`
- `"github.com/cboone/jm/internal/types"` -> `"github.com/cboone/fm/internal/types"`

**Files:**
- `main.go`
- `cmd/root.go`, `cmd/archive.go`, `cmd/dryrun.go`, `cmd/flag.go`, `cmd/list.go`, `cmd/mailboxes.go`, `cmd/mark-read.go`, `cmd/move.go`, `cmd/read.go`, `cmd/search.go`, `cmd/session.go`, `cmd/spam.go`, `cmd/unflag.go`
- `cmd/list_test.go`, `cmd/dryrun_test.go`
- `internal/client/client.go`, `internal/client/email.go`, `internal/client/mailbox.go`
- `internal/client/client_test.go`, `internal/client/email_test.go`, `internal/client/safety_test.go`
- `internal/output/json.go`, `internal/output/text.go`
- `internal/output/json_test.go`, `internal/output/text_test.go`
- `internal/types/types_test.go`

### Step 3: CLI command name and descriptions (`cmd/root.go`)

**File:** `cmd/root.go`

- Line 24: `Use: "jm"` -> `Use: "fm"`
- Line 25: `Short: "JMAP Mail -- a safe, read-oriented CLI for JMAP email (Fastmail)"` -> `Short: "Fastmail Mail -- a safe, read-oriented CLI for Fastmail email via JMAP"`
- Lines 26-28: Update Long description:
  ```
  `fm is a command-line tool for reading, searching, and triaging Fastmail email
  via the JMAP protocol. It connects to Fastmail (or any JMAP server) and
  provides read, search, archive, and spam operations.`
  ```

### Step 4: Config directory path (`cmd/root.go`)

**File:** `cmd/root.go`

- Line 45: `"config file (default: ~/.config/jm/config.yaml)"` -> `"config file (default: ~/.config/fm/config.yaml)"`
- Line 84: `filepath.Join(home, ".config", "jm")` -> `filepath.Join(home, ".config", "fm")`
- Line 110: `"Fix the syntax in ~/.config/jm/config.yaml or use --config"` -> `"Fix the syntax in ~/.config/fm/config.yaml or use --config"`

### Step 5: Environment variable prefix and flag help text (`cmd/root.go`)

**File:** `cmd/root.go`

- Line 46: `"JMAP bearer token"` -> `"Fastmail API token"`
- Line 47: `"JMAP session endpoint"` -> `"Fastmail session endpoint"`
- Line 49: `"JMAP account ID (auto-detected if blank)"` -> `"Fastmail account ID (auto-detected if blank)"`
- Line 91: `viper.SetEnvPrefix("JMAP")` -> `viper.SetEnvPrefix("FM")`

### Step 6: Error message hints -- `JMAP_TOKEN` -> `FM_TOKEN`

**File:** `cmd/root.go`

- Line 117: `"no token configured; set JMAP_TOKEN, --token, or token in config file"` -> `"no token configured; set FM_TOKEN, --token, or token in config file"`

**All cmd files with auth error hints** (same string in each):
- `"Check your token in JMAP_TOKEN or config file"` -> `"Check your token in FM_TOKEN or config file"`

Files: `cmd/archive.go:20`, `cmd/flag.go:19`, `cmd/list.go:40`, `cmd/mailboxes.go:16`, `cmd/mark-read.go:19`, `cmd/move.go:24`, `cmd/read.go:19`, `cmd/search.go:76`, `cmd/session.go:16`, `cmd/spam.go:20`, `cmd/unflag.go:19`

### Step 7: Test helpers referencing `JMAP_TOKEN` in Go tests

**File:** `internal/types/types_test.go`
- Line 362: `Hint: "check JMAP_TOKEN"` -> `Hint: "check FM_TOKEN"`
- Line 378: `if result["hint"] != "check JMAP_TOKEN"` -> `if result["hint"] != "check FM_TOKEN"`

**File:** `internal/output/text_test.go`
- Line 525: `"set JMAP_TOKEN"` -> `"set FM_TOKEN"`
- Line 530: `"Hint: set JMAP_TOKEN"` -> `"Hint: set FM_TOKEN"`

### Step 8: Build config

**File:** `Makefile`
- Line 1: `BINARY := jm` -> `BINARY := fm`
- Line 18: Comment mentions `JMAP_TOKEN and JMAP_LIVE_TESTS` -> `FM_TOKEN and FM_LIVE_TESTS`

**File:** `.gitignore`
- Line 5: `/jm` -> `/fm`

**File:** `.goreleaser.yml`
- Line 4: `id: jm` -> `id: fm`
- Line 6: `binary: jm` -> `binary: fm`
- Line 18: `-X github.com/cboone/jm/cmd.version` -> `-X github.com/cboone/fm/cmd.version`
- Line 22: `- jm` -> `- fm`
- Line 51: `homepage: "https://github.com/cboone/jm"` -> `homepage: "https://github.com/cboone/fm"`
- Line 52: `description: "Safe, read-oriented CLI for JMAP email"` -> `description: "Safe, read-oriented CLI for Fastmail email via JMAP"`
- Line 55: `shell_output("#{bin}/jm --version")` -> `shell_output("#{bin}/fm --version")`

### Step 9: GitHub instructions files

**File:** `.github/go.instructions.md`
- Line 10: `jm search` -> `fm search`
- Line 16: `jm` -> `fm` (two instances in CLI test references)

### Step 10: Documentation

**File:** `README.md` -- Full rewrite of references:
- Title: `# jm` -> `# fm`
- All `jm` command examples -> `fm`
- `brew install cboone/tap/jm` -> `brew install cboone/tap/fm`
- `go install github.com/cboone/jm@latest` -> `go install github.com/cboone/fm@latest`
- `~/.config/jm/config.yaml` -> `~/.config/fm/config.yaml`
- `JMAP_TOKEN` -> `FM_TOKEN`, `JMAP_SESSION_URL` -> `FM_SESSION_URL`, `JMAP_FORMAT` -> `FM_FORMAT`, `JMAP_ACCOUNT_ID` -> `FM_ACCOUNT_ID`
- Description line: "A safe, read-oriented CLI for JMAP email (Fastmail)" -> "A safe, read-oriented CLI for Fastmail email via JMAP"
- Keep JMAP where it refers to the protocol itself (e.g., JMAP scopes, JMAP session endpoint URL, "JMAP protocol")

**File:** `docs/CLI-REFERENCE.md` -- Full rewrite of references:
- Title: `# jm CLI Reference` -> `# fm CLI Reference`
- All `jm` command examples -> `fm`
- Global flags table: `JMAP_TOKEN` -> `FM_TOKEN`, etc.
- `~/.config/jm/config.yaml` -> `~/.config/fm/config.yaml`
- Error reference table: update hints mentioning `~/.config/jm/` and `JMAP_TOKEN`
- Cobra validation section: `"jm"` -> `"fm"`
- Flag descriptions in tables: "JMAP session endpoint" -> "Fastmail session endpoint", etc. (mirrors Step 5 changes)
- Keep "JMAP" only where it refers to the protocol itself (e.g., "JMAP method error" error code description)

**File:** `docs/CLAUDE-CODE-GUIDE.md` -- Full rewrite of references:
- Title: `# Using jm with Claude Code` -> `# Using fm with Claude Code`
- All `jm` command examples -> `fm`
- `go install github.com/cboone/jm@latest` -> `go install github.com/cboone/fm@latest`
- `JMAP_TOKEN` -> `FM_TOKEN`
- `~/.config/jm/config.yaml` -> `~/.config/fm/config.yaml`
- CLAUDE.md snippet section: update all `jm` -> `fm` and `JMAP_TOKEN` -> `FM_TOKEN`
- Keep "JMAP" where it refers to the protocol

**File:** `.env.example`
- `JMAP_TOKEN` -> `FM_TOKEN`
- `JMAP_SESSION_URL` -> `FM_SESSION_URL`
- `JMAP_FORMAT` -> `FM_FORMAT`
- `JMAP_ACCOUNT_ID` -> `FM_ACCOUNT_ID`
- Keep JMAP scope URNs unchanged (`urn:ietf:params:jmap:*`) because they are protocol identifiers

### Step 11: Scrut CLI test files

All test files reference the binary as `$TESTDIR/../jm` and expect output containing `jm`. These all change to `fm`.

**File:** `tests/help.md`
- Title: `# jm help output` -> `# fm help output`
- All `$TESTDIR/../jm` -> `$TESTDIR/../fm`
- Expected output: `jm is a command-line tool for reading, searching, and triaging email` -> `fm is a command-line tool for reading, searching, and triaging Fastmail email`
- Expected output: `jm [command]` -> `fm [command]`
- Expected output: `"jm [command] --help"` -> `"fm [command] --help"`
- All `jm session`, `jm mailboxes`, etc. in Usage lines -> `fm session`, `fm mailboxes`, etc.

**File:** `tests/errors.md`
- Title: `# jm error handling` -> `# fm error handling`
- All `$TESTDIR/../jm` -> `$TESTDIR/../fm`
- All `JMAP_TOKEN` -> `FM_TOKEN`, `JMAP_SESSION_URL` -> `FM_SESSION_URL`, `JMAP_FORMAT` -> `FM_FORMAT`, `JMAP_ACCOUNT_ID` -> `FM_ACCOUNT_ID`
- Expected output strings: `"set JMAP_TOKEN"` -> `"set FM_TOKEN"`, `"Check your token in JMAP_TOKEN"` -> `"Check your token in FM_TOKEN"`

**File:** `tests/flags.md`
- Title: `# jm flag handling` -> `# fm flag handling`
- All `$TESTDIR/../jm` -> `$TESTDIR/../fm`
- Version output: `jm version *` -> `fm version *`
- All `JMAP_TOKEN` -> `FM_TOKEN`, `JMAP_SESSION_URL` -> `FM_SESSION_URL`, `JMAP_FORMAT` -> `FM_FORMAT`, `JMAP_ACCOUNT_ID` -> `FM_ACCOUNT_ID`
- Expected output: `"set JMAP_TOKEN"` -> `"set FM_TOKEN"`, `"Check your token in JMAP_TOKEN"` -> `"Check your token in FM_TOKEN"`
- Section title: `## JMAP_FORMAT env var` -> `## FM_FORMAT env var`
- Section title: `## JMAP_TOKEN env var` -> `## FM_TOKEN env var`
- Section title: `## JMAP_SESSION_URL env var` -> `## FM_SESSION_URL env var`

**File:** `tests/arguments.md`
- Title: `# jm argument validation` -> `# fm argument validation`
- All `$TESTDIR/../jm` -> `$TESTDIR/../fm`
- Expected output: `unknown command "delete" for "jm"` -> `unknown command "delete" for "fm"`
- All `JMAP_TOKEN` -> `FM_TOKEN`, `JMAP_SESSION_URL` -> `FM_SESSION_URL`, `JMAP_FORMAT` -> `FM_FORMAT`, `JMAP_ACCOUNT_ID` -> `FM_ACCOUNT_ID`

**File:** `tests/live.md`
- Title: `# jm live integration tests` -> `# fm live integration tests`
- All `$TESTDIR/../jm` -> `$TESTDIR/../fm`
- Precondition guard: `$JMAP_LIVE_TESTS` -> `$FM_LIVE_TESTS`, `$JMAP_TOKEN` -> `$FM_TOKEN`
- All `JMAP_LIVE_TESTS` -> `FM_LIVE_TESTS`
- `JMAP_TOKEN` -> `FM_TOKEN`

### Step 12: go.sum regeneration

After changing `go.mod`, run `go mod tidy` to regenerate `go.sum` with the new module path.

## Files NOT changed

- `docs/plans/done/*.md` -- historical records, left as-is
- `.github/ci.instructions.md` -- no jm/JMAP references
- `.github/go-mod.instructions.md` -- no jm/JMAP references
- `.github/markdown.instructions.md` -- no jm/JMAP references (mentions PLAN.md only)
- `.github/workflows/ci.yml` -- no jm/JMAP references
- `.github/workflows/release.yml` -- no jm/JMAP references
- `internal/` Go source files -- no string literal "jm" references (only import paths, covered in Step 2)
- Error code strings like `"jmap_error"` -- these are internal error codes referencing the protocol, not the tool name. Leave unchanged.

## JMAP references that remain as-is

These are protocol references, not tool branding:

- `"jmap_error"` error code strings in Go source
- JMAP scope URNs: `urn:ietf:params:jmap:core`, `urn:ietf:params:jmap:mail`
- Default session URL: `https://api.fastmail.com/jmap/session`
- Go code comments referencing the JMAP protocol (e.g., "JMAP session info", "JMAP server", "JMAP client")
- `.github/go.instructions.md` references to JMAP protocol behavior
- Phrase "via JMAP" or "via the JMAP protocol" in descriptions (kept as protocol reference)

## Execution order

1. Steps 1-7 (Go code changes) -- do these first so the code compiles
2. Run `go mod tidy` (Step 12)
3. Run `go build -o fm .` to verify compilation
4. Steps 8-11 (config, docs, tests)
5. Verify

## Verification

1. `go build -o fm .` -- binary builds successfully
2. `go vet ./...` -- no issues
3. `go test ./...` -- unit tests pass
4. `make test-cli` -- scrut integration tests pass (after binary name change)
5. `./fm --version` -- prints `fm version dev`
6. `./fm --help` -- shows Fastmail-branded description with `fm` command name
7. Spot-check `./fm session` (with FM_TOKEN set) to confirm env var prefix works
