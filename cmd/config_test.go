package cmd

import (
	"encoding/json"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"testing"
)

func TestCredentialCommandConfig(t *testing.T) {
	t.Parallel()

	const canonical = "credential_command: 'exit 41'\n"
	const alias = "credential-command: 'exit 42'\n"
	const conflict = "failed to read config: credential_command and credential-command have different values; use one spelling"
	tests := []struct {
		name, config, env, flag, code, message string
	}{
		{name: "underscore", config: canonical, code: "authentication_failed", message: "credential command failed: exit status 41"},
		{name: "hyphen", config: alias, code: "authentication_failed", message: "credential command failed: exit status 42"},
		{name: "same values", config: canonical + "credential-command: 'exit 41'\n", code: "authentication_failed", message: "credential command failed: exit status 41"},
		{name: "conflict", config: canonical + alias, code: "config_error", message: conflict},
		{name: "empty and nonempty conflict", config: "credential_command: ''\n" + alias, code: "config_error", message: conflict},
		{name: "null canonical and nonempty alias", config: "credential_command: null\n" + alias, code: "config_error", message: conflict},
		{name: "nonempty canonical and null alias", config: canonical + "credential-command:\n", code: "config_error", message: conflict},
		{name: "conflict with flag and environment", config: canonical + alias, env: "exit 43", flag: "exit 44", code: "config_error", message: conflict},
		{name: "environment overrides file", config: alias, env: "exit 43", code: "authentication_failed", message: "credential command failed: exit status 43"},
		{name: "flag overrides environment and file", config: alias, env: "exit 43", flag: "exit 44", code: "authentication_failed", message: "credential command failed: exit status 44"},
		{name: "hyphenated environment key ignored", code: "authentication_failed", message: "credential command failed: exit status 1\ncredential access disabled in CLI tests"},
	}

	stubs, err := filepath.Abs("../tests/credential-stubs")
	if err != nil {
		t.Fatal(err)
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			dir := t.TempDir()
			path := filepath.Join(dir, "config.yaml")
			if err := os.WriteFile(path, []byte(tt.config), 0o600); err != nil {
				t.Fatal(err)
			}

			args := []string{"-test.run=^TestCredentialCommandConfigProcess$", "--", "--config", path, "session"}
			if tt.flag != "" {
				args = append(args, "--credential-command", tt.flag)
			}
			child := exec.CommandContext(t.Context(), os.Args[0], args...)
			child.Env = []string{
				"HOME=" + dir,
				"PATH=" + stubs + ":/usr/bin:/bin",
				"FM_TEST_CREDENTIAL_STUBS=" + stubs,
				"FM_SESSION_URL=http://127.0.0.1:1/session",
				"FM_CREDENTIAL-COMMAND=exit 45",
			}
			if tt.env != "" {
				child.Env = append(child.Env, "FM_CREDENTIAL_COMMAND="+tt.env)
			}

			out, err := child.CombinedOutput()
			var exitErr *exec.ExitError
			if !errors.As(err, &exitErr) || exitErr.ExitCode() != 1 {
				t.Fatalf("expected a CLI failure, got %v: %s", err, out)
			}
			var got struct {
				Code    string `json:"error"`
				Message string `json:"message"`
			}
			if err := json.Unmarshal(out, &got); err != nil {
				t.Fatalf("decode CLI error: %v: %s", err, out)
			}
			if got.Code != tt.code || got.Message != tt.message {
				t.Errorf("got %s %q, want %s %q", got.Code, got.Message, tt.code, tt.message)
			}
		})
	}
}

func TestCredentialCommandConfigProcess(t *testing.T) {
	stubs := os.Getenv("FM_TEST_CREDENTIAL_STUBS")
	if stubs == "" {
		t.Skip("CLI test subprocess only")
	}
	for _, name := range []string{"security", "secret-tool"} {
		got, err := exec.LookPath(name)
		if err != nil || got != filepath.Join(stubs, name) {
			t.Fatalf("credential fallback stub not selected for %s: %v", name, err)
		}
	}
	separator := slices.Index(os.Args, "--")
	if separator < 0 {
		t.Fatal("missing CLI arguments")
	}
	args := os.Args[separator+1:]
	if len(args) < 3 || args[0] != "--config" || filepath.Dir(args[1]) != os.Getenv("HOME") ||
		!strings.HasSuffix(args[1], "/config.yaml") {
		t.Fatal("refusing to read config outside the test directory")
	}
	rootCmd.SetArgs(args)
	if err := Execute(); err != nil {
		os.Exit(1)
	}
	os.Exit(0)
}
