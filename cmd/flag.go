package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/cboone/fm/internal/client"
	"github.com/cboone/fm/internal/types"
)

var flagCmd = &cobra.Command{
	Use:   "flag [email-id...]",
	Short: "Flag emails (set the $flagged keyword)",
	Args:  cobra.ArbitraryArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := validateIDsOrFilters(cmd, args); err != nil {
			return err
		}

		colorStr, _ := cmd.Flags().GetString("color")

		var color *client.FlagColor
		if colorStr != "" {
			parsedColor, err := client.ParseFlagColor(colorStr)
			if err != nil {
				return exitError("general_error", err.Error(),
					fmt.Sprintf("Valid colors: %s", strings.Join(client.ValidColorNames(), ", ")))
			}
			color = &parsedColor
		}

		c, err := newClient()
		if err != nil {
			return exitError("authentication_failed", err.Error(),
				"Check your credential command or the token it returns")
		}

		ids, err := resolveEmailIDs(cmd, args, c)
		if err != nil {
			return err
		}

		dryRun, _ := cmd.Flags().GetBool("dry-run")
		if dryRun {
			return dryRunPreview(c, ids, "flag", nil)
		}

		var succeeded, errors []string
		if color != nil {
			succeeded, errors = c.SetFlaggedWithColor(ids, *color)
		} else {
			succeeded, errors = c.SetFlagged(ids)
		}

		result := types.MoveResult{
			Matched:   len(ids),
			Processed: len(succeeded) + len(errors),
			Failed:    len(errors),
			Flagged:   succeeded,
			Errors:    errors,
		}

		if err := formatter().Format(os.Stdout, result); err != nil {
			return err
		}

		if len(errors) > 0 {
			return exitError("partial_failure", "one or more emails failed to flag", "")
		}

		return nil
	},
}

func init() {
	flagCmd.Flags().StringP("color", "c", "", "flag color: red, orange, yellow, green, blue, purple, gray")
	flagCmd.Flags().BoolP("dry-run", "n", false, "preview affected emails without making changes")
	addFilterFlags(flagCmd)
	rootCmd.AddCommand(flagCmd)
}
