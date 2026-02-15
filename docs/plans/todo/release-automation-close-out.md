# Close out release automation (issue #38)

## Context

Issue #38 tracks GoReleaser-based release automation and Homebrew tap publishing. The core implementation (`.goreleaser.yml`, `.github/workflows/release.yml`, version variable with ldflags injection) was merged to main in earlier commits. This branch has no new commits yet.

Three tasks remain to fully close the issue:
1. Document the version strategy (`--version` output contract)
2. Set the `HOMEBREW_TAP_TOKEN` repository secret
3. Cut the first SemVer tag (`v0.2.0`)

## Plan

### 1. Document version strategy

Add a "Versioning" section to `README.md` (after the "Install" section) documenting:
- `fm --version` outputs `fm version <semver>` (e.g., `fm version 0.2.0`)
- Development builds show `fm version dev`
- Releases follow SemVer, injected at build time via GoReleaser ldflags
- Tagged releases (`v*`) trigger automated GitHub Releases and Homebrew formula updates

Add `--version` to the Global Flags table in `docs/CLI-REFERENCE.md`.

**Files modified:**
- `README.md`
- `docs/CLI-REFERENCE.md`

### 2. Set `HOMEBREW_TAP_TOKEN` repository secret

Guide the user through creating a fine-grained GitHub PAT:
- Go to https://github.com/settings/personal-access-tokens/new
- Scope: repository access to `cboone/homebrew-tap` only
- Permission: Contents (read and write)
- Then run: `gh secret set HOMEBREW_TAP_TOKEN` (will prompt for the value)

### 3. Commit, push, create PR, merge

- Commit the documentation changes
- Push and create a PR referencing issue #38
- After merge, proceed to tagging

### 4. Cut `v0.2.0` tag on main

After the PR is merged:
- `git checkout main && git pull`
- `git tag -s v0.2.0 -m "v0.2.0"`
- `git push origin v0.2.0`

This triggers the release workflow, which builds binaries, creates the GitHub Release, and pushes the stable Homebrew formula to `cboone/homebrew-tap`.

## Verification

1. `fm --version` outputs `fm version dev` (local build)
2. `gh secret list` shows `HOMEBREW_TAP_TOKEN`
3. After tagging, `gh run list --workflow=release.yml` shows a successful run
4. `gh release view v0.2.0` shows the release with binaries and checksums
5. The formula at `cboone/homebrew-tap/Formula/fm.rb` is updated with a stable URL and sha256
