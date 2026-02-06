package cmd

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/cboone/jm/internal/client"
	"github.com/cboone/jm/internal/output"
)

// ErrSilent is returned by exitError to indicate the error has already been printed.
var ErrSilent = errors.New("error already printed")

var (
	cfgFile string
	rootCmd = &cobra.Command{
		Use:   "jm",
		Short: "JMAP Mail -- a safe, read-oriented CLI for JMAP email (Fastmail)",
		Long: `jm is a command-line tool for reading, searching, and triaging email
via the JMAP protocol. It connects to Fastmail (or any JMAP server) and
provides read, search, archive, and spam operations.

Sending and deleting email are structurally disallowed.`,
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

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default: ~/.config/jm/config.yaml)")
	rootCmd.PersistentFlags().String("token", "", "JMAP bearer token")
	rootCmd.PersistentFlags().String("session-url", "https://api.fastmail.com/jmap/session", "JMAP session endpoint")
	rootCmd.PersistentFlags().String("format", "json", "output format: json or text")
	rootCmd.PersistentFlags().String("account-id", "", "JMAP account ID (auto-detected if blank)")

	for _, bind := range []struct{ key, flag string }{
		{"token", "token"},
		{"session_url", "session-url"},
		{"format", "format"},
		{"account_id", "account-id"},
	} {
		if err := viper.BindPFlag(bind.key, rootCmd.PersistentFlags().Lookup(bind.flag)); err != nil {
			panic(fmt.Sprintf("failed to bind flag %q: %v", bind.flag, err))
		}
	}
}

func initConfig() {
	if cfgFile != "" {
		viper.SetConfigFile(cfgFile)
	} else {
		home, err := os.UserHomeDir()
		if err == nil {
			configDir := filepath.Join(home, ".config", "jm")
			viper.AddConfigPath(configDir)
			viper.SetConfigName("config")
			viper.SetConfigType("yaml")
		}
	}

	viper.SetEnvPrefix("JMAP")
	viper.AutomaticEnv()

	viper.SetDefault("session_url", "https://api.fastmail.com/jmap/session")
	viper.SetDefault("format", "json")

	// Ignore errors if config file doesn't exist.
	viper.ReadInConfig()
}

// newClient creates an authenticated JMAP client from the current config.
func newClient() (*client.Client, error) {
	token := viper.GetString("token")
	if token == "" {
		return nil, fmt.Errorf("no token configured; set JMAP_TOKEN, --token, or token in config file")
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
	formatter().FormatError(os.Stderr, code, message, hint)
	return ErrSilent
}
