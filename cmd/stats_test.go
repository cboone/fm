package cmd

import "testing"

func TestStatsCmd_FlaggedAndUnflaggedMutuallyExclusive(t *testing.T) {
	rootCmd.SetArgs([]string{"stats", "--flagged", "--unflagged"})
	err := rootCmd.Execute()
	if err == nil {
		t.Fatal("expected error when both --flagged and --unflagged are set")
	}
}
