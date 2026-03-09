package cmd

import (
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/cboone/fm/internal/types"
)

var sieveDeleteCmd = &cobra.Command{
	Use:   "delete <script-id>",
	Short: "Delete a sieve script",
	Long: `Delete a sieve script by ID.

Active scripts cannot be deleted. Use 'sieve deactivate' first.`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		dryRun, _ := cmd.Flags().GetBool("dry-run")
		if dryRun {
			result := types.SieveDryRunResult{
				Operation: "delete",
				Script:    args[0],
			}
			return formatter().Format(os.Stdout, result)
		}

		c, err := newClient()
		if err != nil {
			return exitError("authentication_failed", err.Error(),
				"Check your credential command or the token it returns")
		}

		result, err := c.DeleteSieveScript(args[0])
		if err != nil {
			if strings.Contains(err.Error(), "deactivate it first") {
				return exitError("forbidden_operation", err.Error(),
					"Use 'fm sieve deactivate' before deleting")
			}
			return exitError("jmap_error", err.Error(), "")
		}

		return formatter().Format(os.Stdout, result)
	},
}

func init() {
	sieveDeleteCmd.Flags().BoolP("dry-run", "n", false, "preview without making changes")
	sieveCmd.AddCommand(sieveDeleteCmd)
}
