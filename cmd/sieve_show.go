package cmd

import (
	"os"
	"strings"

	"github.com/spf13/cobra"
)

var sieveShowCmd = &cobra.Command{
	Use:   "show <script-id>",
	Short: "Show a sieve script's metadata and content",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newClient()
		if err != nil {
			return exitError("authentication_failed", err.Error(),
				"Check your credential command or the token it returns")
		}

		result, err := c.GetSieveScript(args[0])
		if err != nil {
			if strings.Contains(err.Error(), "not found") {
				return exitError("not_found", err.Error(), "")
			}
			return exitError("jmap_error", err.Error(), "")
		}

		return formatter().Format(os.Stdout, result)
	},
}

func init() {
	sieveCmd.AddCommand(sieveShowCmd)
}
