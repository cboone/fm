package cmd

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/spf13/cobra"

	"github.com/cboone/fm/internal/types"
)

func TestDraftWarningsOutput(t *testing.T) {
	previous := config.Get("format")
	t.Cleanup(func() { config.Set("format", previous) })
	for _, format := range []string{"json", "text"} {
		t.Run(format, func(t *testing.T) {
			config.Set("format", format)
			var stdout, stderr bytes.Buffer
			command := &cobra.Command{}
			command.SetOut(&stdout)
			command.SetErr(&stderr)
			result := types.DraftResult{ID: "synthetic", Mode: "reply", Warnings: []string{"synthetic fallback"}}
			if err := formatDraftResult(command, result); err != nil {
				t.Fatal(err)
			}
			if format == "json" {
				var decoded map[string]json.RawMessage
				if err := json.Unmarshal(stdout.Bytes(), &decoded); err != nil {
					t.Fatal(err)
				}
				var warnings []string
				if err := json.Unmarshal(decoded["warnings"], &warnings); err != nil {
					t.Fatalf("literal warnings array missing or invalid: %v", err)
				}
				if len(warnings) != 1 || warnings[0] != "synthetic fallback" || stderr.Len() != 0 {
					t.Fatalf("JSON warnings=%v stderr=%q", warnings, stderr.String())
				}
			} else if stderr.String() != "Warning: synthetic fallback\n" || strings.Contains(stdout.String(), "fallback") || !strings.Contains(stdout.String(), "Draft created: synthetic") {
				t.Fatalf("text stdout=%q stderr=%q", stdout.String(), stderr.String())
			}
			stdout.Reset()
			stderr.Reset()
			result.Warnings = nil
			if err := formatDraftResult(command, result); err != nil {
				t.Fatal(err)
			}
			if strings.Contains(stdout.String(), "warnings") || stderr.Len() != 0 {
				t.Fatalf("no warnings: stdout=%q stderr=%q", stdout.String(), stderr.String())
			}
			if format == "json" {
				var decoded map[string]json.RawMessage
				if err := json.Unmarshal(stdout.Bytes(), &decoded); err != nil {
					t.Fatal(err)
				}
				if _, present := decoded["warnings"]; present {
					t.Fatal("literal warnings key must be omitted when empty")
				}
			}
		})
	}
}
