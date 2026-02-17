package cmd

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// TestCLIReferenceCoverage verifies that every flag registered on key commands
// appears in docs/CLI-REFERENCE.md. This catches silent drift when flags are
// added or removed in code without updating the documentation.
func TestCLIReferenceCoverage(t *testing.T) {
	doc, err := os.ReadFile("../docs/CLI-REFERENCE.md")
	if err != nil {
		t.Fatalf("failed to read CLI-REFERENCE.md: %v", err)
	}
	content := string(doc)

	commands := map[string]*cobra.Command{
		"session":   sessionCmd,
		"list":      listCmd,
		"search":    searchCmd,
		"read":      readCmd,
		"mailboxes": mailboxesCmd,
		"move":      moveCmd,
		"archive":   archiveCmd,
		"spam":      spamCmd,
		"mark-read": markReadCmd,
		"flag":      flagCmd,
		"unflag":    unflagCmd,
	}

	for name, cmd := range commands {
		// Extract the section for this command from the docs.
		section := extractCommandSection(content, name)
		if section == "" {
			t.Errorf("docs/CLI-REFERENCE.md has no section for %q", name)
			continue
		}

		cmd.Flags().VisitAll(func(f *pflag.Flag) {
			if f.Name == "help" {
				return // Cobra's built-in --help is never documented
			}

			flagRef := fmt.Sprintf("`--%s`", f.Name)
			if !strings.Contains(section, flagRef) {
				t.Errorf("command %q: flag --%s is registered in code but missing from docs/CLI-REFERENCE.md", name, f.Name)
			}

			if f.Shorthand != "" {
				shortRef := fmt.Sprintf("`-%s`", f.Shorthand)
				if !strings.Contains(section, shortRef) {
					t.Errorf("command %q: short flag -%s (for --%s) is registered in code but missing from docs/CLI-REFERENCE.md", name, f.Shorthand, f.Name)
				}
			}
		})
	}
}

// TestHelpTestCoverage verifies that every flag registered on key commands
// appears in the scrut help snapshot tests. This catches drift when flags are
// added in code without updating the help test expectations.
func TestHelpTestCoverage(t *testing.T) {
	doc, err := os.ReadFile("../tests/help.md")
	if err != nil {
		t.Fatalf("failed to read tests/help.md: %v", err)
	}
	content := string(doc)

	commands := map[string]*cobra.Command{
		"session":   sessionCmd,
		"list":      listCmd,
		"search":    searchCmd,
		"read":      readCmd,
		"mailboxes": mailboxesCmd,
		"move":      moveCmd,
		"archive":   archiveCmd,
		"spam":      spamCmd,
		"mark-read": markReadCmd,
		"flag":      flagCmd,
		"unflag":    unflagCmd,
	}

	for name, cmd := range commands {
		section := extractHelpCommandSection(content, name)
		if section == "" {
			t.Errorf("tests/help.md has no help block for %q", name)
			continue
		}

		cmd.Flags().VisitAll(func(f *pflag.Flag) {
			if f.Name == "help" {
				return
			}

			flagRef := fmt.Sprintf("--%s", f.Name)
			if !strings.Contains(section, flagRef) {
				t.Errorf("command %q: flag --%s is registered in code but missing from tests/help.md", name, f.Name)
			}

			if f.Shorthand != "" {
				shortRef := fmt.Sprintf("-%s,", f.Shorthand)
				if !strings.Contains(section, shortRef) {
					t.Errorf("command %q: short flag -%s (for --%s) is registered in code but missing from tests/help.md", name, f.Shorthand, f.Name)
				}
			}
		})
	}
}

// TestGlobalFlagsCoverage verifies that every persistent flag registered on
// rootCmd appears in the Global Flags table of docs/CLI-REFERENCE.md.
func TestGlobalFlagsCoverage(t *testing.T) {
	doc, err := os.ReadFile("../docs/CLI-REFERENCE.md")
	if err != nil {
		t.Fatalf("failed to read CLI-REFERENCE.md: %v", err)
	}
	content := string(doc)

	// Extract the Global Flags section (between "## Global Flags" and the next "---").
	const heading = "## Global Flags"
	start := strings.Index(content, heading)
	if start == -1 {
		t.Fatal("docs/CLI-REFERENCE.md has no Global Flags section")
	}
	rest := content[start+len(heading):]
	end := strings.Index(rest, "\n---")
	if end == -1 {
		end = len(rest)
	}
	section := rest[:end]

	rootCmd.PersistentFlags().VisitAll(func(f *pflag.Flag) {
		if f.Name == "help" {
			return
		}

		flagRef := fmt.Sprintf("`--%s`", f.Name)
		if !strings.Contains(section, flagRef) {
			t.Errorf("global flag --%s is registered on rootCmd but missing from the Global Flags table in docs/CLI-REFERENCE.md", f.Name)
		}
	})

	// Cobra adds --version automatically when rootCmd.Version is set.
	if rootCmd.Version != "" {
		if !strings.Contains(section, "`--version`") {
			t.Errorf("global flag --version is registered on rootCmd but missing from the Global Flags table in docs/CLI-REFERENCE.md")
		}
	}
}

// extractCommandSection returns the docs content between a "### <name>"
// heading and the next "---" separator (or "## " heading).
func extractCommandSection(doc, name string) string {
	heading := "### " + name
	start := strings.Index(doc, heading)
	if start == -1 {
		return ""
	}
	rest := doc[start+len(heading):]

	// Find the end of this section: next "---" or "## " heading.
	end := len(rest)
	for _, sep := range []string{"\n---", "\n## "} {
		if idx := strings.Index(rest, sep); idx != -1 && idx < end {
			end = idx
		}
	}
	return rest[:end]
}

// extractHelpCommandSection returns the output block for a command in tests/help.md,
// starting from "$ $TESTDIR/../fm <name> --help" until the closing code fence.
func extractHelpCommandSection(doc, name string) string {
	command := "$ $TESTDIR/../fm " + name + " --help"
	start := strings.Index(doc, command)
	if start == -1 {
		return ""
	}

	rest := doc[start+len(command):]
	end := strings.Index(rest, "\n```")
	if end == -1 {
		return rest
	}

	return rest[:end]
}
