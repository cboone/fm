package cmd

import (
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"

	"github.com/cboone/fm/internal/client"
	"github.com/cboone/fm/internal/types"
)

var sieveCreateCmd = &cobra.Command{
	Use:   "create --name <name> [flags]",
	Short: "Create a new sieve script",
	Long: `Create a new sieve script on the server.

Two modes are supported:

Template mode (generate a script from flags):
  fm sieve create --name "Block sender" --from "spam@example.com" --action junk
  fm sieve create --name "Block domain" --from-domain "example.com" --action junk

Stdin mode (provide raw sieve content):
  echo 'keep;' | fm sieve create --name "Custom" --script-stdin

By default, new scripts are created inactive. Use --activate to activate
on creation. Only one script can be active at a time; activating a new
script deactivates the currently active one.`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		name, _ := cmd.Flags().GetString("name")
		if name == "" {
			return exitError("general_error", "required flag \"name\" not set",
				"Provide a name with --name")
		}

		content, err := resolveSieveCreateContent(cmd)
		if err != nil {
			return exitError("general_error", err.Error(), "")
		}

		activate, _ := cmd.Flags().GetBool("activate")
		dryRun, _ := cmd.Flags().GetBool("dry-run")

		if dryRun {
			result := types.SieveDryRunResult{
				Operation: "create",
				Script:    name,
				Content:   content,
			}

			// Optionally validate server-side during dry run.
			c, clientErr := newClient()
			if clientErr == nil {
				valResult, valErr := c.ValidateSieveScript(content)
				if valErr == nil {
					result.Valid = &valResult.Valid
				}
			}

			return formatter().Format(os.Stdout, result)
		}

		c, err := newClient()
		if err != nil {
			return exitError("authentication_failed", err.Error(),
				"Check your credential command or the token it returns")
		}

		result, err := c.CreateSieveScript(name, content, activate)
		if err != nil {
			return exitError("jmap_error", err.Error(), "")
		}

		return formatter().Format(os.Stdout, result)
	},
}

func init() {
	sieveCreateCmd.Flags().String("name", "", "name for the new script (required)")
	sieveCreateCmd.Flags().String("from", "", "match sender email address (template mode)")
	sieveCreateCmd.Flags().String("from-domain", "", "match sender domain (template mode)")
	sieveCreateCmd.Flags().String("action", "", "action: junk, discard, keep, or fileinto (template mode)")
	sieveCreateCmd.Flags().String("fileinto", "", "target mailbox for fileinto action (template mode)")
	sieveCreateCmd.Flags().Bool("script-stdin", false, "read raw sieve script from stdin")
	sieveCreateCmd.Flags().Bool("activate", false, "activate the script immediately after creation")
	sieveCreateCmd.Flags().BoolP("dry-run", "n", false, "preview the generated script without creating it")
	sieveCmd.AddCommand(sieveCreateCmd)
}

// resolveSieveCreateContent determines the script content from either template
// flags or stdin.
func resolveSieveCreateContent(cmd *cobra.Command) (string, error) {
	scriptStdin, _ := cmd.Flags().GetBool("script-stdin")
	from, _ := cmd.Flags().GetString("from")
	fromDomain, _ := cmd.Flags().GetString("from-domain")
	action, _ := cmd.Flags().GetString("action")

	hasTemplate := from != "" || fromDomain != "" || action != ""

	if hasTemplate && scriptStdin {
		return "", fmt.Errorf("template flags (--from, --from-domain, --action) and --script-stdin are mutually exclusive")
	}
	if !hasTemplate && !scriptStdin {
		return "", fmt.Errorf("provide either template flags (--from/--from-domain + --action) or --script-stdin")
	}

	if scriptStdin {
		data, err := io.ReadAll(os.Stdin)
		if err != nil {
			return "", fmt.Errorf("reading script from stdin: %w", err)
		}
		return string(data), nil
	}

	fileinto, _ := cmd.Flags().GetString("fileinto")
	return client.GenerateSieveScript(client.SieveTemplateOptions{
		From:       from,
		FromDomain: fromDomain,
		Action:     action,
		FileInto:   fileinto,
	})
}
