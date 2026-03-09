package client

import (
	"fmt"
	"strings"
)

// SieveTemplateOptions configures sieve script generation from CLI flags.
type SieveTemplateOptions struct {
	From       string // exact sender address match
	FromDomain string // sender domain match
	Action     string // junk, discard, keep, or fileinto
	FileInto   string // target mailbox name (required when Action is "fileinto")
}

// GenerateSieveScript produces a complete sieve script from template options.
func GenerateSieveScript(opts SieveTemplateOptions) (string, error) {
	if err := validateTemplateOptions(opts); err != nil {
		return "", err
	}

	var b strings.Builder

	action, needsFileinto := sieveAction(opts)
	if needsFileinto {
		b.WriteString("require [\"fileinto\"];\n\n")
	}

	condition := sieveCondition(opts)
	_, _ = fmt.Fprintf(&b, "if %s {\n", condition)
	_, _ = fmt.Fprintf(&b, "    %s\n", action)
	b.WriteString("    stop;\n")
	b.WriteString("}\n")

	return b.String(), nil
}

func validateTemplateOptions(opts SieveTemplateOptions) error {
	if opts.From == "" && opts.FromDomain == "" {
		return fmt.Errorf("either --from or --from-domain is required")
	}
	if opts.From != "" && opts.FromDomain != "" {
		return fmt.Errorf("--from and --from-domain are mutually exclusive")
	}
	switch opts.Action {
	case "junk", "discard", "keep":
		// valid
	case "fileinto":
		if opts.FileInto == "" {
			return fmt.Errorf("--fileinto is required when --action is fileinto")
		}
	case "":
		return fmt.Errorf("--action is required")
	default:
		return fmt.Errorf("unsupported action %q: use junk, discard, keep, or fileinto", opts.Action)
	}
	return nil
}

func sieveCondition(opts SieveTemplateOptions) string {
	if opts.From != "" {
		return fmt.Sprintf("address :is \"from\" %q", opts.From)
	}
	return fmt.Sprintf("address :domain :is \"from\" %q", opts.FromDomain)
}

func sieveAction(opts SieveTemplateOptions) (action string, needsFileinto bool) {
	switch opts.Action {
	case "junk":
		return "fileinto \"Junk\";", true
	case "discard":
		return "discard;", false
	case "keep":
		return "keep;", false
	case "fileinto":
		return fmt.Sprintf("fileinto %q;", opts.FileInto), true
	default:
		return "keep;", false
	}
}
