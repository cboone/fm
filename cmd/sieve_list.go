package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

var sieveListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all sieve scripts",
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newClient()
		if err != nil {
			return exitError("authentication_failed", err.Error(),
				"Check your credential command or the token it returns")
		}

		result, err := c.ListSieveScripts()
		if err != nil {
			return exitError("jmap_error", err.Error(), "")
		}

		return formatter().Format(os.Stdout, result)
	},
}

func init() {
	sieveCmd.AddCommand(sieveListCmd)
}
