package cmd

import (
	"fmt"
	"io"
	"net/mail"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/cboone/fm/internal/client"
	"github.com/cboone/fm/internal/types"
)

var draftCmd = &cobra.Command{
	Use:   "draft [flags]",
	Short: "Create a draft email in the Drafts mailbox",
	Long: `Create a draft email for later review and sending from Fastmail.

Supports four composition modes:
  - New message: provide --to, --subject, and --body/--body-stdin
  - Reply: provide --reply-to <email-id> and --body/--body-stdin
  - Reply-all: provide --reply-all <email-id> and --body/--body-stdin
  - Forward: provide --forward <email-id>, --to, and --body/--body-stdin

The draft is placed in the Drafts mailbox with $draft and $seen keywords.
It is NOT sent. Review and send from Fastmail.`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		mode, err := determineDraftMode(cmd)
		if err != nil {
			return exitError("general_error", err.Error(), "")
		}

		body, err := readDraftBody(cmd)
		if err != nil {
			return exitError("general_error", err.Error(), "")
		}

		subject, _ := cmd.Flags().GetString("subject")
		html, _ := cmd.Flags().GetBool("html")

		to, err := parseAddressFlag(cmd, "to")
		if err != nil {
			return exitError("general_error", err.Error(), "use RFC 5322 format: \"Name <email>\" or just email")
		}
		cc, err := parseAddressFlag(cmd, "cc")
		if err != nil {
			return exitError("general_error", err.Error(), "use RFC 5322 format: \"Name <email>\" or just email")
		}
		bcc, err := parseAddressFlag(cmd, "bcc")
		if err != nil {
			return exitError("general_error", err.Error(), "use RFC 5322 format: \"Name <email>\" or just email")
		}

		if err := validateDraftFlags(mode, to, subject); err != nil {
			return exitError("general_error", err.Error(), "")
		}

		var originalID string
		switch mode {
		case client.DraftModeReply:
			originalID, _ = cmd.Flags().GetString("reply-to")
		case client.DraftModeReplyAll:
			originalID, _ = cmd.Flags().GetString("reply-all")
		case client.DraftModeForward:
			originalID, _ = cmd.Flags().GetString("forward")
		}

		c, err := newClient()
		if err != nil {
			return exitError("authentication_failed", err.Error(),
				"Check your credential command or the token it returns")
		}

		result, err := c.CreateDraft(client.DraftOptions{
			Mode:       mode,
			To:         to,
			CC:         cc,
			BCC:        bcc,
			Subject:    subject,
			Body:       body,
			HTML:       html,
			OriginalID: originalID,
		})
		if err != nil {
			if _, ok := err.(*client.ErrForbidden); ok {
				return exitError("forbidden_operation", err.Error(), "")
			}
			if strings.Contains(err.Error(), "not found") {
				return exitError("not_found", err.Error(), "")
			}
			return exitError("jmap_error", err.Error(), "")
		}

		return formatter().Format(os.Stdout, result)
	},
}

func init() {
	draftCmd.Flags().StringSlice("to", nil, "recipient addresses (RFC 5322)")
	draftCmd.Flags().StringSlice("cc", nil, "CC addresses (RFC 5322)")
	draftCmd.Flags().StringSlice("bcc", nil, "BCC addresses (RFC 5322)")
	draftCmd.Flags().String("subject", "", "subject line")
	draftCmd.Flags().String("body", "", "message body")
	draftCmd.Flags().Bool("body-stdin", false, "read body from stdin")
	draftCmd.Flags().String("reply-to", "", "email ID to reply to")
	draftCmd.Flags().String("reply-all", "", "email ID to reply-all to")
	draftCmd.Flags().String("forward", "", "email ID to forward")
	draftCmd.Flags().Bool("html", false, "treat body as HTML")

	rootCmd.AddCommand(draftCmd)
}

// determineDraftMode checks the mode flags and returns the mode.
func determineDraftMode(cmd *cobra.Command) (client.DraftMode, error) {
	replyTo, _ := cmd.Flags().GetString("reply-to")
	replyAll, _ := cmd.Flags().GetString("reply-all")
	forward, _ := cmd.Flags().GetString("forward")

	count := 0
	if replyTo != "" {
		count++
	}
	if replyAll != "" {
		count++
	}
	if forward != "" {
		count++
	}
	if count > 1 {
		return "", fmt.Errorf("--reply-to, --reply-all, and --forward are mutually exclusive")
	}

	switch {
	case replyTo != "":
		return client.DraftModeReply, nil
	case replyAll != "":
		return client.DraftModeReplyAll, nil
	case forward != "":
		return client.DraftModeForward, nil
	default:
		return client.DraftModeNew, nil
	}
}

// readDraftBody reads the body from either --body or --body-stdin.
func readDraftBody(cmd *cobra.Command) (string, error) {
	bodyStr, _ := cmd.Flags().GetString("body")
	bodyStdin, _ := cmd.Flags().GetBool("body-stdin")

	if bodyStr != "" && bodyStdin {
		return "", fmt.Errorf("--body and --body-stdin are mutually exclusive")
	}
	if bodyStr == "" && !bodyStdin {
		return "", fmt.Errorf("either --body or --body-stdin is required")
	}

	if bodyStdin {
		data, err := io.ReadAll(os.Stdin)
		if err != nil {
			return "", fmt.Errorf("reading body from stdin: %w", err)
		}
		return string(data), nil
	}

	return bodyStr, nil
}

// parseAddressFlag parses an address flag value into types.Address slices.
// Supports RFC 5322 "Name <email>" and bare "email" formats.
func parseAddressFlag(cmd *cobra.Command, flag string) ([]types.Address, error) {
	values, _ := cmd.Flags().GetStringSlice(flag)
	if len(values) == 0 {
		return nil, nil
	}

	var addrs []types.Address
	for _, v := range values {
		v = strings.TrimSpace(v)
		if v == "" {
			continue
		}
		parsed, err := mail.ParseAddress(v)
		if err != nil {
			return nil, fmt.Errorf("invalid %s address %q: %w", flag, v, err)
		}
		addrs = append(addrs, types.Address{Name: parsed.Name, Email: parsed.Address})
	}
	return addrs, nil
}

// validateDraftFlags checks mode-specific required flags.
func validateDraftFlags(mode client.DraftMode, to []types.Address, subject string) error {
	switch mode {
	case client.DraftModeNew:
		if len(to) == 0 {
			return fmt.Errorf("--to is required for new drafts")
		}
		if subject == "" {
			return fmt.Errorf("--subject is required for new drafts")
		}
	case client.DraftModeForward:
		if len(to) == 0 {
			return fmt.Errorf("--to is required for forwarded drafts")
		}
	}
	return nil
}
