# Fix Homebrew shell completions

## Context

After `brew install cboone/tap/fm` (v0.3.0), shell completions are not available. Cobra auto-registers an `fm completion [bash|zsh|fish]` subcommand that works correctly, but the Homebrew cask does not generate or install completion scripts. Users must manually run `fm completion zsh > ...` and source the output, which is a poor experience.

The fix: pre-generate completion scripts during the goreleaser build, bundle them in release archives, and declare them in the `homebrew_casks.completions` field so Homebrew installs them automatically.

## Changes

### 1. Create `scripts/completions.sh`

New file. Builds the binary, generates completion scripts for bash/zsh/fish, cleans up.

```bash
#!/usr/bin/env bash
set -euo pipefail

rm -rf completions
mkdir completions

go build -o fm .
./fm completion bash > completions/fm.bash
./fm completion zsh  > completions/fm.zsh
./fm completion fish > completions/fm.fish
rm fm
```

### 2. Update `.goreleaser.yml`

Three additions to the existing config:

**a) Add `before.hooks`** to run the completion generation script:

```yaml
before:
  hooks:
    - go mod tidy
    - bash scripts/completions.sh
```

**b) Add `files` to `archives`** so completion scripts are bundled in tar.gz/zip:

```yaml
archives:
  - # ... existing config ...
    files:
      - completions/*
```

**c) Add `completions` to `homebrew_casks`** so Homebrew knows where to find them:

```yaml
homebrew_casks:
  - # ... existing config ...
    completions:
      bash: completions/fm.bash
      zsh: completions/fm.zsh
      fish: completions/fm.fish
```

### 3. Update `.gitignore`

Add `completions/` under the "Build output" section (alongside `bin/` and `dist/`).

### 4. Update `.github/goreleaser.instructions.md`

Add a bullet documenting the completions setup so future agent sessions don't remove or misunderstand it.

## Files to modify

- `scripts/completions.sh` (new)
- `.goreleaser.yml`
- `.gitignore`
- `.github/goreleaser.instructions.md`

## What stays the same

- `cmd/root.go`: no changes needed, Cobra auto-registers `completion`
- `cmd/docs_drift_test.go`: already skips `completion` in coverage checks
- `docs/CLI-REFERENCE.md`: completions are an install-time concern, not a CLI docs concern
- `tests/help.md`: already includes `completion` in root help output
- `homebrew_casks` (not switching to deprecated `brews`): correct per goreleaser v2.10+
- `xattr` quarantine hook: still needed for unsigned binary

## Verification

1. **Validate config**: `goreleaser check`
2. **Test completion generation**: `bash scripts/completions.sh && ls completions/` (should show `fm.bash`, `fm.zsh`, `fm.fish`)
3. **Dry-run build**: `goreleaser release --snapshot --clean`
4. **Inspect archives**: verify `completions/` directory is present in archives under `dist/`
5. **Inspect generated cask**: `cat dist/homebrew/Casks/fm.rb`, confirm completion stanzas are present
6. **Manual smoke test**: `source completions/fm.zsh && fm <TAB>` should show subcommands
