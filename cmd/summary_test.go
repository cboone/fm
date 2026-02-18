package cmd

import "testing"

func TestSummaryCmd_FlaggedAndUnflaggedMutuallyExclusive(t *testing.T) {
	rootCmd.SetArgs([]string{"summary", "--flagged", "--unflagged"})
	err := rootCmd.Execute()
	if err == nil {
		t.Fatal("expected error when both --flagged and --unflagged are set")
	}
}

func TestSummaryCmd_LimitMustBePositive(t *testing.T) {
	rootCmd.SetArgs([]string{"summary", "--limit", "0"})
	err := rootCmd.Execute()
	if err == nil {
		t.Fatal("expected error when --limit is 0")
	}
}

func TestSummaryCmd_NoPositionalArgs(t *testing.T) {
	rootCmd.SetArgs([]string{"summary", "extra"})
	err := rootCmd.Execute()
	if err == nil {
		t.Fatal("expected error when positional args are provided")
	}
}
