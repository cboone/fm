package cmd

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/mail"
	"net/url"
	"os"
	"strings"

	"git.sr.ht/~rockorager/go-jmap"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/cboone/fm/internal/client"
	"github.com/cboone/fm/internal/types"
	"github.com/cboone/fm/internal/unsubscribe"
)

type oneClickMailbox interface {
	NewestSenderEmail(string) (*unsubscribe.Candidate, error)
}

type oneClickSetupError struct {
	credential bool
	err        error
}

func (e *oneClickSetupError) Error() string { return "one-click client setup failed" }
func (e *oneClickSetupError) Unwrap() error { return e.err }

var (
	oneClickOpenMailbox = openOneClickMailbox
	oneClickHTTPClient  = unsubscribe.NewHTTPClient
)

func openOneClickMailbox() (oneClickMailbox, error) {
	token, err := resolveToken()
	if err != nil {
		return nil, &oneClickSetupError{credential: true, err: err}
	}
	c, err := client.New(viper.GetString("session_url"), token, viper.GetString("account_id"))
	if err != nil {
		return nil, &oneClickSetupError{err: err}
	}
	return c, nil
}

func oneClickSenders(cmd *cobra.Command, args []string) ([]string, error) {
	oneClick, _ := cmd.Flags().GetBool("one-click")
	draft, _ := cmd.Flags().GetBool("draft")
	confirm, _ := cmd.Flags().GetBool("confirm")
	dryRun, _ := cmd.Flags().GetBool("dry-run")
	if !oneClick || draft || len(args) != 0 || (confirm && dryRun) {
		return nil, fmt.Errorf("use --one-click --sender addresses, without IDs, --draft or conflicting --confirm/--dry-run")
	}
	for _, name := range filterFlagNames {
		if cmd.Flags().Changed(name) {
			return nil, fmt.Errorf("--one-click does not accept filter flags")
		}
	}
	inputs, _ := cmd.Flags().GetStringArray("sender")
	if len(inputs) == 0 {
		return nil, fmt.Errorf("--one-click requires at least one --sender")
	}
	var senders []string
	seen := make(map[string]bool)
	for _, input := range inputs {
		address, err := mail.ParseAddress(input)
		if err != nil || address.Name != "" || address.Address != input || strings.ContainsAny(input, "*?\r\n\t \"<>\\[]") || strings.Count(input, "@") != 1 {
			return nil, fmt.Errorf("--sender requires a plain mailbox address")
		}
		key := strings.ToLower(input)
		if !seen[key] {
			seen[key] = true
			senders = append(senders, input)
		}
	}
	return senders, nil
}

func runOneClickCommand(cmd *cobra.Command, args []string) error {
	senders, err := oneClickSenders(cmd, args)
	if err != nil {
		return exitError("general_error", err.Error(), "")
	}
	// Authentication headers are meaningful only within the Fastmail trust boundary.
	if viper.GetString("session_url") != "https://api.fastmail.com/jmap/session" {
		return exitError("general_error", "one-click requires the Fastmail session endpoint", "")
	}
	c, err := oneClickOpenMailbox()
	if err != nil {
		var setupErr *oneClickSetupError
		if errors.As(err, &setupErr) && setupErr.credential {
			return exitError("authentication_failed", "credential command failed or returned no token",
				"Check FM_CREDENTIAL_COMMAND or --credential-command, including its timeout; no session request was made")
		}
		failure := oneClickSessionFailure(err)
		return exitError(failure.Error, failure.Message, failure.Hint)
	}
	httpClient := oneClickHTTPClient()
	defer httpClient.CloseIdleConnections()
	confirm, _ := cmd.Flags().GetBool("confirm")
	results, failed := oneClickResults(cmd.Context(), senders, confirm, c.NewestSenderEmail, httpClient)
	if err := formatter().Format(os.Stdout, results); err != nil {
		return err
	}
	if failed {
		return exitError("partial_failure", "one or more sender operations failed; inspect per-sender results", "do not retry uncertain requests automatically")
	}
	return nil
}

func oneClickSessionFailure(err error) types.AppError {
	failure := types.AppError{Error: "authentication_failed", Message: "Fastmail session discovery failed", Hint: "Check token permissions and --account-id"}
	if errors.Is(err, client.ErrNoPrimaryAccount) {
		failure.Message, failure.Hint = "Fastmail session has no primary mail account", "Check --account-id and token permissions"
		return failure
	}
	cause, hint := oneClickProtocolCause(err)
	failure.Message += ": " + cause
	failure.Hint = hint
	return failure
}

func oneClickProtocolCause(err error) (string, string) {
	var methodErr *jmap.MethodError
	if errors.As(err, &methodErr) {
		switch methodErr.Type {
		case "serverFail":
			return "JMAP serverFail", "check Fastmail service availability and try a fresh preview later"
		case "forbidden":
			return "JMAP forbidden", "check account and token permissions"
		case "accountNotFound", "accountNotSupportedByMethod", "accountReadOnly":
			return "JMAP " + methodErr.Type, "check --account-id and account capabilities"
		case "invalidArguments", "unknownMethod", "unsupportedFilter", "unsupportedSort":
			return "JMAP " + methodErr.Type, "check CLI compatibility with Fastmail"
		default:
			return "JMAP method error", "check Fastmail service availability and CLI compatibility"
		}
	}
	var requestErr *jmap.RequestError
	if errors.As(err, &requestErr) && requestErr.Status >= 400 && requestErr.Status <= 599 {
		hint := "check Fastmail service availability and try a fresh preview later"
		if requestErr.Status == 401 || requestErr.Status == 403 {
			hint = "check token and account permissions"
		}
		return fmt.Sprintf("HTTP %d", requestErr.Status), hint
	}
	var networkErr *url.Error
	if errors.As(err, &networkErr) {
		return "connection failure", "check Fastmail connectivity"
	}
	if errors.Is(err, client.ErrSenderResponseInvalid) {
		return "invalid metadata response", "check Fastmail service availability and CLI compatibility"
	}
	return "unclassified transport or protocol failure", "check Fastmail service availability and try a fresh preview later"
}

func oneClickResults(ctx context.Context, senders []string, confirm bool, selectEmail func(string) (*unsubscribe.Candidate, error), httpClient *http.Client) ([]types.OneClickResult, bool) {
	results := make([]types.OneClickResult, 0, len(senders))
	failed := false
	for _, sender := range senders {
		result, senderFailed := oneClickResult(ctx, sender, confirm, selectEmail, httpClient)
		results = append(results, result)
		failed = failed || senderFailed
	}
	return results, failed
}

func oneClickResult(ctx context.Context, sender string, confirm bool, selectEmail func(string) (*unsubscribe.Candidate, error), httpClient *http.Client) (types.OneClickResult, bool) {
	result := types.OneClickResult{Sender: sender, Status: "skipped"}
	message, err := selectEmail(sender)
	if err != nil {
		result.Status = "error"
		if errors.Is(err, client.ErrSenderSelectionChanged) {
			result.Reason = "sender selection changed; preview again"
		} else {
			operation := "sender metadata lookup"
			var lookupErr *client.SenderLookupError
			if errors.As(err, &lookupErr) {
				switch lookupErr.Operation {
				case "query":
					operation = "sender query"
				case "get":
					operation = "sender metadata read"
				}
			}
			cause, hint := oneClickProtocolCause(err)
			result.Reason = operation + " failed: " + cause + "; " + hint
		}
		return result, true
	}
	if message == nil {
		result.Reason = "no exact sender message found"
		return result, false
	}
	if len(message.From) != 1 || !strings.EqualFold(message.From[0], sender) {
		result.Reason = "ambiguous or nonmatching From address"
		return result, false
	}
	result.EmailID = message.ID
	result.ReceivedAt = message.ReceivedAt.Format("2006-01-02T15:04:05Z07:00")
	target, err := unsubscribe.Target(message.Headers)
	if err != nil {
		result.Reason = err.Error()
		return result, false
	}
	_, domain, _ := strings.Cut(sender, "@")
	if err := unsubscribe.Authenticated(message.Headers, domain); err != nil {
		result.Reason = err.Error()
		return result, false
	}
	result.Host = target.Hostname()
	result.Status = "previewed"
	if !confirm {
		return result, false
	}
	result.HTTPStatus, err = unsubscribe.Post(ctx, httpClient, target)
	result.Status = "http_status"
	if err != nil {
		result.Status, result.Reason = "error", err.Error()
	}
	return result, err != nil || result.HTTPStatus < 200 || result.HTTPStatus >= 300
}
