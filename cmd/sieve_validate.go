package cmd

import (
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"
)

var sieveValidateCmd = &cobra.Command{
	Use:   "validate",
	Short: "Validate sieve script syntax without storing it",
	Long: `Validate sieve script syntax on the server without creating a script.

Provide the script content via --script or --script-stdin.`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		content, err := readSieveContent(cmd)
		if err != nil {
			return exitError("general_error", err.Error(), "")
		}

		c, err := newClient()
		if err != nil {
			return exitError("authentication_failed", err.Error(),
				"Check your credential command or the token it returns")
		}

		result, err := c.ValidateSieveScript(content)
		if err != nil {
			return exitError("jmap_error", err.Error(), "")
		}

		return formatter().Format(os.Stdout, result)
	},
}

func init() {
	sieveValidateCmd.Flags().String("script", "", "sieve script content as a string")
	sieveValidateCmd.Flags().Bool("script-stdin", false, "read sieve script from stdin")
	sieveCmd.AddCommand(sieveValidateCmd)
}

// readSieveContent reads script content from --script or --script-stdin.
func readSieveContent(cmd *cobra.Command) (string, error) {
	script, _ := cmd.Flags().GetString("script")
	scriptStdin, _ := cmd.Flags().GetBool("script-stdin")

	if script != "" && scriptStdin {
		return "", fmt.Errorf("--script and --script-stdin are mutually exclusive")
	}
	if script == "" && !scriptStdin {
		return "", fmt.Errorf("either --script or --script-stdin is required")
	}

	if scriptStdin {
		data, err := io.ReadAll(os.Stdin)
		if err != nil {
			return "", fmt.Errorf("reading script from stdin: %w", err)
		}
		return string(data), nil
	}

	return script, nil
}
