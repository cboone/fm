package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

var sessionCmd = &cobra.Command{
	Use:   "session",
	Short: "Display JMAP session info (verify connectivity and auth)",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newClient()
		if err != nil {
			return exitError("authentication_failed", err.Error(),
				"Check your credential command or the token it returns")
		}

		info := c.SessionInfo()
		return formatter().Format(os.Stdout, info)
	},
}

func init() {
	rootCmd.AddCommand(sessionCmd)
}
