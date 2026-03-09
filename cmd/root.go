package cmd

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/cboone/fm/internal/client"
	"github.com/cboone/fm/internal/output"
)

// ErrSilent is returned by exitError to indicate the error has already been printed.
var ErrSilent = errors.New("error already printed")

var (
	cfgFile       string
	initConfigErr error
	version       = "dev"
	rootCmd       = &cobra.Command{
		Use:   "fm",
		Short: "Fastmail Mail -- a safe, read-oriented CLI for Fastmail email via JMAP",
		Long: `fm is a command-line tool for reading, searching, triaging, and drafting Fastmail
email via the JMAP protocol. It connects to Fastmail (or any JMAP server) and
provides read, search, archive, spam, and draft operations.

Sending and deleting email are structurally disallowed.`,
		Version:       version,
		SilenceUsage:  true,
		SilenceErrors: true,
	}
)

// Execute runs the root command.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default: ~/.config/fm/config.yaml)")
	rootCmd.PersistentFlags().String("credential-command", "", "shell command that prints the API token to stdout")
	rootCmd.PersistentFlags().String("session-url", "https://api.fastmail.com/jmap/session", "Fastmail session endpoint")
	rootCmd.PersistentFlags().String("format", "json", "output format: json or text")
	rootCmd.PersistentFlags().String("account-id", "", "Fastmail account ID (auto-detected if blank)")

	for _, bind := range []struct{ key, flag string }{
		{"credential_command", "credential-command"},
		{"session_url", "session-url"},
		{"format", "format"},
		{"account_id", "account-id"},
	} {
		if err := viper.BindPFlag(bind.key, rootCmd.PersistentFlags().Lookup(bind.flag)); err != nil {
			panic(fmt.Sprintf("failed to bind flag %q: %v", bind.flag, err))
		}
	}

	rootCmd.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
		if initConfigErr != nil {
			return exitError("config_error", "failed to read config: "+initConfigErr.Error(), configErrorHint())
		}
		format := viper.GetString("format")
		if format != "json" && format != "text" {
			return exitError("general_error",
				fmt.Sprintf("unsupported output format: %q", format),
				"supported formats: json, text")
		}
		return nil
	}
}

func initConfig() {
	initConfigErr = nil

	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		home, err := os.UserHomeDir()
		if err == nil {
			configDir := filepath.Join(home, ".config", "fm")
			viper.AddConfigPath(configDir)
			viper.SetConfigName("config")
			viper.SetConfigType("yaml")
		}
	}

	viper.SetEnvPrefix("FM")
	viper.AutomaticEnv()

	viper.SetDefault("session_url", "https://api.fastmail.com/jmap/session")
	viper.SetDefault("format", "json")

	if err := viper.ReadInConfig(); err != nil {
		var notFound viper.ConfigFileNotFoundError
		if errors.As(err, &notFound) {
			return
		}
		initConfigErr = err
	}
}

func configErrorHint() string {
	if cfgFile != "" {
		return "Fix the syntax in " + cfgFile + " or choose another file with --config"
	}
	return "Fix the syntax in ~/.config/fm/config.yaml or use --config"
}

// defaultCredentialCommand returns a platform-specific credential command
// that retrieves the token from the OS keychain.
func defaultCredentialCommand() string {
	switch runtime.GOOS {
	case "darwin":
		return "security find-generic-password -s fm -a fastmail -w"
	case "linux":
		return "secret-tool lookup service fm"
	default:
		return ""
	}
}

// resolveToken executes the configured credential command and returns the token.
func resolveToken() (string, error) {
	credCmd := viper.GetString("credential_command")
	if credCmd == "" {
		credCmd = defaultCredentialCommand()
	}
	if credCmd == "" {
		return "", fmt.Errorf("no credential command configured; set FM_CREDENTIAL_COMMAND, --credential-command, or credential_command in config file")
	}

	var stdout, stderr bytes.Buffer
	cmd := exec.Command("sh", "-c", credCmd)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		detail := strings.TrimSpace(stderr.String())
		if detail != "" {
			return "", fmt.Errorf("credential command failed: %w\n%s", err, detail)
		}
		return "", fmt.Errorf("credential command failed: %w", err)
	}

	token := strings.TrimSpace(stdout.String())
	if token == "" {
		return "", fmt.Errorf("credential command returned empty token")
	}
	return token, nil
}

// newClient creates an authenticated JMAP client from the current config.
func newClient() (*client.Client, error) {
	token, err := resolveToken()
	if err != nil {
		return nil, err
	}
	sessionURL := viper.GetString("session_url")
	accountID := viper.GetString("account_id")

	return client.New(sessionURL, token, accountID)
}

// formatter returns the configured output formatter.
func formatter() output.Formatter {
	return output.New(viper.GetString("format"))
}

// exitError writes a structured error to stderr and returns ErrSilent
// to signal that the error has already been printed.
func exitError(code string, message string, hint string) error {
	if err := formatter().FormatError(os.Stderr, code, message, hint); err != nil {
		fmt.Fprintf(os.Stderr, "error [%s]: %s\n", code, message)
		if hint != "" {
			fmt.Fprintf(os.Stderr, "hint: %s\n", hint)
		}
	}
	return ErrSilent
}
