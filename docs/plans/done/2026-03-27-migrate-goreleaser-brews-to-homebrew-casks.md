# Migrate GoReleaser config from brews to homebrew_casks

## Context

GoReleaser v2.10+ deprecated the `brews` configuration section in favor of `homebrew_casks`. The current `.goreleaser.yml` uses `brews`, which produces a deprecation warning during releases. The CI reusable workflow (`cboone/gh-actions`) already uses GoReleaser v2.14.3, so the new config is fully supported.

Issue: #62

## Changes

### 1. Replace `brews` with `homebrew_casks` in `.goreleaser.yml`

**File:** `.goreleaser.yml` (lines 45-56)

Replace:

```yaml
brews:
  - repository:
      owner: cboone
      name: homebrew-tap
      token: "{{ .Env.HOMEBREW_TAP_TOKEN }}"
    directory: Formula
    homepage: "https://github.com/cboone/fm"
    description: "Safe, read-oriented CLI for Fastmail email via JMAP"
    license: MIT
    test: |
      assert_match version.to_s, shell_output("#{bin}/fm --version")
```

With:

```yaml
homebrew_casks:
  - repository:
      owner: cboone
      name: homebrew-tap
      token: "{{ .Env.HOMEBREW_TAP_TOKEN }}"
    homepage: "https://github.com/cboone/fm"
    description: "Safe, read-oriented CLI for Fastmail email via JMAP"
    url:
      verified: "github.com/cboone/fm/"
    hooks:
      post:
        install: |
          system_command "/usr/bin/xattr",
                         args: ["-dr", "com.apple.quarantine", "#{staged_path}"]
```

Field-by-field decisions:

- **`directory`**: Omitted. `Casks` is the default for `homebrew_casks`, so specifying it is unnecessary.
- **`license`**: Dropped. Not a valid field in the cask DSL. The license is declared in the repo's `LICENSE` file.
- **`test`**: Dropped. Homebrew casks do not support a `test` stanza. Binary testing remains covered by the project's own CI (scrut tests).
- **`url.verified`**: Added. Since `fm` is an unsigned binary, this satisfies Homebrew's URL audit compliance by confirming the download source domain matches the project.
- **`hooks.post.install`**: Added. Removes macOS quarantine attributes (`com.apple.quarantine`) from the unsigned binary after installation, preventing Gatekeeper "unidentified developer" dialogs.

### 2. No change to README.md

The install command `brew install cboone/tap/fm` works identically for both formulas and casks when using a custom tap. No update needed.

### 3. Follow-up work (separate repo, after next release)

Tracked in cboone/homebrew-tap#14. These are issue #62 steps 5-6, to be done in the `cboone/homebrew-tap` repo after cutting a release with the new config:

1. Verify the generated cask in `Casks/fm.rb` works: `brew install cboone/tap/fm`
2. Delete `Formula/fm.rb`
3. Add `tap_migrations.json`: `{ "fm": "cask/fm" }`

This must happen after a release so the cask exists before the formula is removed.

## Verification

1. Run `goreleaser check` to validate the YAML schema (if goreleaser is installed locally)
2. Inspect the diff to confirm only the `brews` block changed
3. After merging and tagging a release, verify the cask is generated in `cboone/homebrew-tap/Casks/`
