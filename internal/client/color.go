package client

import (
	"fmt"
	"slices"
	"strings"

	"git.sr.ht/~rockorager/go-jmap"
	"git.sr.ht/~rockorager/go-jmap/mail/email"
)

// FlagColor represents a flag color per RFC 9979. The three keywords $MailFlagBit0,
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

var colorNames = [...]string{
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
	return slices.Clone(colorNames[:])
}

// ParseFlagColor parses a color name (case-insensitive) into a FlagColor.
func ParseFlagColor(s string) (FlagColor, error) {
	lower := strings.ToLower(s)
	for color, name := range colorNames {
		if name == lower {
			return FlagColor(color), nil
		}
	}
	return 0, fmt.Errorf("invalid color %q; valid colors: %s", s, strings.Join(ValidColorNames(), ", "))
}

func (c FlagColor) String() string {
	if c >= 0 && int(c) < len(colorNames) {
		return colorNames[c]
	}
	return fmt.Sprintf("unknown(%d)", int(c))
}

type flagColorKeyword struct {
	name      string
	lowercase string
	patchPath string
}

func newFlagColorKeyword(name string) flagColorKeyword {
	return flagColorKeyword{name, strings.ToLower(name), "keywords/" + name}
}

var flagColorKeywords = [...]flagColorKeyword{
	newFlagColorKeyword("$MailFlagBit0"),
	newFlagColorKeyword("$MailFlagBit1"),
	newFlagColorKeyword("$MailFlagBit2"),
}

// Patch returns a jmap.Patch that sets the three $MailFlagBit keywords
// for this color. Bits that should be set use true; bits that should be
// cleared use nil (JMAP patch deletion).
func (c FlagColor) Patch() jmap.Patch {
	patch := make(jmap.Patch, len(flagColorKeywords))
	for bit, keyword := range flagColorKeywords {
		var value any
		if c&(1<<bit) != 0 {
			value = true
		}
		patch[keyword.patchPath] = value
	}
	return patch
}

// clearColorPatch returns a jmap.Patch that clears all three $MailFlagBit keywords.
func clearColorPatch() jmap.Patch {
	return FlagColorRed.Patch()
}

func decodeFlagColor(keywords map[string]bool) *string {
	if !keywords["$flagged"] {
		return nil
	}
	var color FlagColor
	for bit, keyword := range flagColorKeywords {
		// JMAP servers return keywords in lowercase.
		if keywords[keyword.lowercase] || keywords[keyword.name] {
			color |= 1 << bit
		}
	}
	if int(color) >= len(colorNames) {
		return nil
	}
	name := colorNames[color]
	return &name
}

func (c FlagColor) filter(base email.Filter) email.Filter {
	conditions := []email.Filter{base}
	for bit, keyword := range flagColorKeywords {
		condition := &email.FilterCondition{}
		if c&(1<<bit) != 0 {
			condition.HasKeyword = keyword.lowercase
		} else {
			condition.NotKeyword = keyword.lowercase
		}
		conditions = append(conditions, condition)
	}
	return &email.FilterOperator{
		Operator:   jmap.OperatorAND,
		Conditions: conditions,
	}
}
