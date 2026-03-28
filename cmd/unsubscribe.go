package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/cboone/fm/internal/client"
	"github.com/cboone/fm/internal/types"
	"github.com/cboone/fm/internal/unsubscribe"
)

var unsubscribeCmd = &cobra.Command{
	Use:   "unsubscribe [email-id]",
	Short: "Show or act on the List-Unsubscribe header of an email",
	Long: `Extract and display the List-Unsubscribe header from an email.

Shows the unsubscribe mechanism: mailto, url, both, or none.
With --draft, creates a draft email for mailto-based unsubscribe.

Accepts a single email ID as argument, or use filter flags to select
an email (e.g. --from sender@example.com).`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := validateIDsOrFilters(cmd, args); err != nil {
			return err
		}

		c, err := newClient()
		if err != nil {
			return exitError("authentication_failed", err.Error(),
				"Check your token in FM_TOKEN or config file")
		}

		ids, err := resolveEmailIDs(cmd, args, c)
		if err != nil {
			return err
		}
		emailID := ids[0]

		detail, err := c.ReadEmail(emailID, false, false)
		if err != nil {
			return exitError(readErrorCode(err), err.Error(), "")
		}

		parsed := unsubscribe.Parse(detail.ListUnsubscribe, detail.ListUnsubscribePost)

		result := types.UnsubscribeResult{
			EmailID:   emailID,
			Mechanism: parsed.Mechanism,
			OneClick:  parsed.OneClick,
			URL:       parsed.URL,
		}
		if parsed.Mailto != nil {
			result.Mailto = parsed.Mailto.Address
			result.Subject = parsed.Mailto.Subject
			result.Body = parsed.Mailto.Body
		}

		draft, _ := cmd.Flags().GetBool("draft")
		if draft {
			if parsed.Mailto == nil {
				hint := "this email only supports URL-based unsubscribe"
				if parsed.Mechanism == unsubscribe.MechanismNone {
					hint = "no List-Unsubscribe header found"
				}
				return exitError("general_error",
					"no mailto unsubscribe address found", hint)
			}

			subject := parsed.Mailto.Subject
			if subject == "" {
				subject = "Unsubscribe"
			}

			result.Subject = subject
			result.Body = parsed.Mailto.Body

			draftResult, draftErr := c.CreateDraft(client.DraftOptions{
				Mode:    client.DraftModeNew,
				To:      []types.Address{{Email: parsed.Mailto.Address}},
				Subject: subject,
				Body:    parsed.Mailto.Body,
			})
			if draftErr != nil {
				if _, ok := draftErr.(*client.ErrForbidden); ok {
					return exitError("forbidden_operation", draftErr.Error(), "")
				}
				if strings.Contains(draftErr.Error(), "not found") {
					return exitError("not_found", draftErr.Error(), "")
				}
				return exitError("jmap_error", fmt.Sprintf("creating unsubscribe draft: %v", draftErr), "")
			}
			result.DraftID = draftResult.ID
		}

		return formatter().Format(os.Stdout, result)
	},
}

func init() {
	unsubscribeCmd.Flags().Bool("draft", false, "create a draft unsubscribe email (mailto only)")
	addFilterFlags(unsubscribeCmd)
	rootCmd.AddCommand(unsubscribeCmd)
}
