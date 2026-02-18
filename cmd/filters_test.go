package cmd

import (
	"errors"
	"testing"

	"github.com/spf13/cobra"
)

func newFilterTestCommand(withDestinationTo bool) *cobra.Command {
	cmd := &cobra.Command{Use: "test"}
	if withDestinationTo {
		cmd.Flags().String("to", "", "target mailbox name or ID (required)")
	}
	addFilterFlags(cmd)
	return cmd
}

func TestHasFilterFlags_IgnoresMoveDestinationTo(t *testing.T) {
	cmd := newFilterTestCommand(true)

	if err := cmd.Flags().Set("to", "Archive"); err != nil {
		t.Fatalf("set --to: %v", err)
	}

	if hasFilterFlags(cmd) {
		t.Fatal("expected destination --to not to count as a filter")
	}

	if err := cmd.Flags().Set("unread", "true"); err != nil {
		t.Fatalf("set --unread: %v", err)
	}

	if !hasFilterFlags(cmd) {
		t.Fatal("expected --unread=true to count as a filter")
	}
}

func TestParseFilterOptions_IgnoresMoveDestinationTo(t *testing.T) {
	cmd := newFilterTestCommand(true)

	if err := cmd.Flags().Set("to", "Archive"); err != nil {
		t.Fatalf("set --to: %v", err)
	}
	if err := cmd.Flags().Set("unread", "true"); err != nil {
		t.Fatalf("set --unread: %v", err)
	}

	opts, err := parseFilterOptions(cmd, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if opts.To != "" {
		t.Fatalf("expected recipient filter To to be empty, got %q", opts.To)
	}
	if !opts.UnreadOnly {
		t.Fatal("expected UnreadOnly=true")
	}
}

func TestHasFilterFlags_IgnoresFalseBooleanFilters(t *testing.T) {
	cmd := newFilterTestCommand(false)

	if err := cmd.Flags().Set("unread", "false"); err != nil {
		t.Fatalf("set --unread: %v", err)
	}

	if hasFilterFlags(cmd) {
		t.Fatal("expected --unread=false not to count as a filter")
	}
}

func TestValidateIDsOrFilters_EmptyStringFilterRejected(t *testing.T) {
	cmd := newFilterTestCommand(false)

	if err := cmd.Flags().Set("subject", ""); err != nil {
		t.Fatalf("set --subject: %v", err)
	}

	err := validateIDsOrFilters(cmd, nil)
	if !errors.Is(err, ErrSilent) {
		t.Fatalf("expected ErrSilent, got %v", err)
	}
}

func TestParseFilterOptions_RecipientToFilterStillWorks(t *testing.T) {
	cmd := newFilterTestCommand(false)

	if err := cmd.Flags().Set("to", "bob@example.com"); err != nil {
		t.Fatalf("set --to: %v", err)
	}

	opts, err := parseFilterOptions(cmd, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if opts.To != "bob@example.com" {
		t.Fatalf("expected To=bob@example.com, got %q", opts.To)
	}
}
