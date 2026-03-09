package cmd

import (
	"os"

	"github.com/spf13/cobra"

	"github.com/cboone/fm/internal/types"
)

var sieveDeactivateCmd = &cobra.Command{
	Use:   "deactivate",
	Short: "Deactivate the currently active sieve script",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		dryRun, _ := cmd.Flags().GetBool("dry-run")
		if dryRun {
			result := types.SieveDryRunResult{
				Operation: "deactivate",
			}
			return formatter().Format(os.Stdout, result)
		}

		c, err := newClient()
		if err != nil {
			return exitError("authentication_failed", err.Error(),
				"Check your credential command or the token it returns")
		}

		result, err := c.DeactivateSieveScript()
		if err != nil {
			return exitError("jmap_error", err.Error(), "")
		}

		return formatter().Format(os.Stdout, result)
	},
}

func init() {
	sieveDeactivateCmd.Flags().BoolP("dry-run", "n", false, "preview without making changes")
	sieveCmd.AddCommand(sieveDeactivateCmd)
}
