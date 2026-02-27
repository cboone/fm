package cmd

import (
	"os"

	"github.com/spf13/cobra"

	"github.com/cboone/fm/internal/types"
)

var sieveActivateCmd = &cobra.Command{
	Use:   "activate <script-id>",
	Short: "Activate a sieve script",
	Long: `Activate a sieve script by ID.

Only one script can be active at a time. Activating a script automatically
deactivates any currently active script.`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		dryRun, _ := cmd.Flags().GetBool("dry-run")
		if dryRun {
			result := types.SieveDryRunResult{
				Operation: "activate",
				Script:    args[0],
			}
			return formatter().Format(os.Stdout, result)
		}

		c, err := newClient()
		if err != nil {
			return exitError("authentication_failed", err.Error(),
				"Check your credential command or the token it returns")
		}

		result, err := c.ActivateSieveScript(args[0])
		if err != nil {
			return exitError("jmap_error", err.Error(), "")
		}

		return formatter().Format(os.Stdout, result)
	},
}

func init() {
	sieveActivateCmd.Flags().BoolP("dry-run", "n", false, "preview without making changes")
	sieveCmd.AddCommand(sieveActivateCmd)
}
