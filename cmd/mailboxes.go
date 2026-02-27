package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

var mailboxesCmd = &cobra.Command{
	Use:   "mailboxes",
	Short: "List all mailboxes (folders/labels) in the account",
	RunE: func(cmd *cobra.Command, args []string) error {
		c, err := newClient()
		if err != nil {
			return exitError("authentication_failed", err.Error(),
				"Check your credential command or the token it returns")
		}

		rolesOnly, _ := cmd.Flags().GetBool("roles-only")
		mailboxes, err := c.ListMailboxes(rolesOnly)
		if err != nil {
			return exitError("jmap_error", err.Error(), "")
		}

		return formatter().Format(os.Stdout, mailboxes)
	},
}

func init() {
	mailboxesCmd.Flags().Bool("roles-only", false, "only show mailboxes with a defined role")
	rootCmd.AddCommand(mailboxesCmd)
}
