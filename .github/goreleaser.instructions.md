---
applyTo: ".goreleaser.yml"
---

- **`homebrew_casks` is correct, not `brews`**: GoReleaser v2.10+ deprecated the `brews` configuration section in favor of `homebrew_casks`. Do not suggest reverting to `brews` or adding a `brews` section alongside `homebrew_casks`.
- **`xattr` quarantine removal in cask hooks is intentional**: The `hooks.post.install` stanza that removes `com.apple.quarantine` is required because the binary is not codesigned or notarized. Homebrew casks are macOS-only, so `/usr/bin/xattr` is always available. Do not suggest removing this hook or flagging it as a Gatekeeper bypass concern.
