package client

import (
	"fmt"
	"strings"

	"git.sr.ht/~rockorager/go-jmap"
)

// FlagColor represents a flag color per the IETF MailFlagBit spec
// (draft-eggert-mailflagcolors-00). The three keywords $MailFlagBit0,
// $MailFlagBit1, and $MailFlagBit2 encode a 3-bit value (0-6) that
// maps to a display color in Apple Mail and Fastmail.
type FlagColor int

const (
	FlagColorRed    FlagColor = 0
	FlagColorOrange FlagColor = 1
	FlagColorYellow FlagColor = 2
	FlagColorGreen  FlagColor = 3
	FlagColorBlue   FlagColor = 4
	FlagColorPurple FlagColor = 5
	FlagColorGray   FlagColor = 6
)

var colorNames = map[FlagColor]string{
	FlagColorRed:    "red",
	FlagColorOrange: "orange",
	FlagColorYellow: "yellow",
	FlagColorGreen:  "green",
	FlagColorBlue:   "blue",
	FlagColorPurple: "purple",
	FlagColorGray:   "gray",
}

// ValidColorNames returns the list of valid color names for use in error messages.
func ValidColorNames() []string {
	return []string{"red", "orange", "yellow", "green", "blue", "purple", "gray"}
}

// ParseFlagColor parses a color name (case-insensitive) into a FlagColor.
func ParseFlagColor(s string) (FlagColor, error) {
	lower := strings.ToLower(s)
	for color, name := range colorNames {
		if name == lower {
			return color, nil
		}
	}
	return 0, fmt.Errorf("invalid color %q; valid colors: %s", s, strings.Join(ValidColorNames(), ", "))
}

func (c FlagColor) String() string {
	if name, ok := colorNames[c]; ok {
		return name
	}
	return fmt.Sprintf("unknown(%d)", int(c))
}

// bit returns true/nil for a single bit position (0, 1, or 2).
func (c FlagColor) bit(pos int) any {
	if int(c)&(1<<pos) != 0 {
		return true
	}
	return nil
}

// Patch returns a jmap.Patch that sets the three $MailFlagBit keywords
// for this color. Bits that should be set use true; bits that should be
// cleared use nil (JMAP patch deletion).
func (c FlagColor) Patch() jmap.Patch {
	return jmap.Patch{
		"keywords/$MailFlagBit0": c.bit(0),
		"keywords/$MailFlagBit1": c.bit(1),
		"keywords/$MailFlagBit2": c.bit(2),
	}
}

// clearColorPatch returns a jmap.Patch that clears all three $MailFlagBit keywords.
func clearColorPatch() jmap.Patch {
	return jmap.Patch{
		"keywords/$MailFlagBit0": nil,
		"keywords/$MailFlagBit1": nil,
		"keywords/$MailFlagBit2": nil,
	}
}
