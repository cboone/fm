package cmd

import (
	"os"

	"github.com/cboone/jm/internal/client"
	"github.com/cboone/jm/internal/types"
)

func dryRunPreview(c *client.Client, ids []string, operation string, dest *types.DestinationInfo) error {
	summaries, notFound, err := c.GetEmailSummaries(ids)
	if err != nil {
		return exitError("jmap_error", err.Error(), "")
	}

	result := types.DryRunResult{
		Operation:   operation,
		Count:       len(summaries),
		Emails:      summaries,
		NotFound:    notFound,
		Destination: dest,
	}

	if err := formatter().Format(os.Stdout, result); err != nil {
		return err
	}

	if len(notFound) > 0 {
		return exitError("partial_failure", "one or more email IDs were not found", "")
	}

	return nil
}
