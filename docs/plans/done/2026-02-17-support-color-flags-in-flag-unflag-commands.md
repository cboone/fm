# Support color flags in flag/unflag commands

## Context

The `flag` command currently only sets the generic `$flagged` keyword. Fastmail and Apple Mail support colored flags via three JMAP/IMAP keywords (`$MailFlagBit0`, `$MailFlagBit1`, `$MailFlagBit2`) defined in the [IETF draft](https://www.ietf.org/archive/id/draft-eggert-mailflagcolors-00.html). These three bits encode a color value 0-6:

| Value | Bit0 | Bit1 | Bit2 | Color  |
|-------|------|------|------|--------|
| 0     | 0    | 0    | 0    | red    |
| 1     | 1    | 0    | 0    | orange |
| 2     | 0    | 1    | 0    | yellow |
| 3     | 1    | 1    | 0    | green  |
| 4     | 0    | 0    | 1    | blue   |
| 5     | 1    | 0    | 1    | purple |
| 6     | 0    | 1    | 1    | gray   |

**Note:** The issue description has green and blue swapped vs. the IETF standard. This plan follows the IETF standard.

**Key spec rules:**
- Color bits SHOULD be ignored if `$flagged` is not set
- Clients SHOULD NOT set color bits unless `$flagged` is being set
- Clients SHOULD clear all 3 color bits when unflagging

## Design

### `flag` command

Add a `--color` / `-c` flag accepting one of: `red`, `orange`, `yellow`, `green`, `blue`, `purple`, `gray`.

Behavior:
- `fm flag M1` — Sets `$flagged` only (unchanged, backward compatible)
- `fm flag --color orange M1` — Sets `$flagged` + sets orange color bits (`$MailFlagBit0`=true, `$MailFlagBit1`=nil, `$MailFlagBit2`=nil)
- `fm flag --color red M1` — Sets `$flagged` + clears all color bits (red = default = no bits)

When `--color` is provided, the JMAP patch includes both `$flagged` and all three `$MailFlagBit` keywords (set to `true` or `nil` as needed for the color).

### `unflag` command

Add a `--color` / `-c` boolean flag.

Behavior:
- `fm unflag M1` — Removes `$flagged` + clears all 3 color bits (per IETF spec recommendation)
- `fm unflag --color M1` — Removes color bits only, leaves `$flagged` intact. Useful for resetting to default red flag without unflagging.

This is a **minor behavior change** to plain `unflag`: it now also clears color bits. This is safe because color bits are meaningless without `$flagged`, and the spec recommends this cleanup.

### Validation

Invalid `--color` values produce a `general_error` with a hint listing valid colors.

## Files to modify

### 1. `internal/client/email.go`

Add a new exported type and constants for flag colors, and new client methods:

```go
// FlagColor represents a flag color per the IETF MailFlagBit spec.
type FlagColor int

const (
    FlagColorNone   FlagColor = -1
    FlagColorRed    FlagColor = 0
    FlagColorOrange FlagColor = 1
    FlagColorYellow FlagColor = 2
    FlagColorGreen  FlagColor = 3
    FlagColorBlue   FlagColor = 4
    FlagColorPurple FlagColor = 5
    FlagColorGray   FlagColor = 6
)

func ParseFlagColor(s string) (FlagColor, error) { ... }
func (c FlagColor) String() string { ... }
func (c FlagColor) Patch() jmap.Patch { ... }  // Returns the 3-bit keyword patch
```

- Add `SetFlaggedWithColor(emailIDs []string, color FlagColor) ([]string, []string)` — Sets `$flagged` + color bits
- Modify `SetUnflagged` to also clear all 3 `$MailFlagBit` keywords
- Add `ClearFlagColor(emailIDs []string) ([]string, []string)` — Clears only color bits (for `unflag --color`)

### 2. `cmd/flag.go`

- Add `--color` / `-c` boolean flag (default: false)
- When set, validate with `ParseFlagColor`, call `SetFlaggedWithColor`
- When not set, call existing `SetFlagged` (backward compatible)

### 3. `cmd/unflag.go`

- Add `--color` / `-c` string flag (default: empty)
- When `--color` is NOT set: call updated `SetUnflagged` (which now also clears color bits)
- When `--color` IS set: call `ClearFlagColor` (only clears color bits, keeps `$flagged`)

### 4. `internal/client/email_test.go`

Add tests for:
- `ParseFlagColor` — valid names, invalid names, case sensitivity
- `FlagColor.Patch()` — verify correct bit patterns for all 7 colors
- `SetFlaggedWithColor` — verify JMAP patch includes `$flagged` + color bits
- Updated `SetUnflagged` — verify it clears all 3 `$MailFlagBit` keywords
- `ClearFlagColor` — verify it clears only color bits, not `$flagged`

### 5. `docs/CLI-REFERENCE.md`

Update the `flag` and `unflag` command sections to document:
- New `--color` / `-c` flag
- Valid color values
- Behavior changes

### 6. `tests/help.md`

Update the flag and unflag help output snapshots to include `--color`.

### 7. `tests/flags.md`

Add CLI tests for:
- `fm flag --color orange M1` (passes flag parsing, fails on auth)
- `fm flag --color invalid M1` (returns validation error)
- `fm unflag --color M1` (passes flag parsing, fails on auth)

## Verification

1. `go build -o /tmp/fm-test .` — Confirm it compiles
2. `go test ./...` — All unit tests pass
3. `make test-cli` — All CLI snapshot tests pass
4. Manual smoke test: `/tmp/fm-test flag --help` shows `--color` flag
5. Manual smoke test: `/tmp/fm-test flag --color invalid M1` shows error with valid colors hint
6. If a live Fastmail account is available, test `flag --color orange <id>` and verify the color appears in Fastmail web UI
