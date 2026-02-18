---
applyTo: "**/*.md"
---

- **PLAN.md documents actual behavior**: PLAN.md is a living design document that reflects the current implementation. Do not flag discrepancies that have already been updated.
- **Exit codes**: All errors exit with code 1. There are no distinct exit codes for different error categories.
- **Error format**: Runtime command errors use structured JSON to stderr. Cobra argument/flag validation errors use Cobra's default plain-text format. This is intentional.
- **Output examples include destination**: The `archive`, `spam`, and `move` command output examples in PLAN.md include the `destination` object. Do not flag missing destination fields.
- **MailboxInfo parent_id uses omitempty**: `parent_id` is omitted from JSON output when empty. PLAN.md examples reflect this. Do not suggest showing `"parent_id": null`.
- **Plan documents are historical records**: Files in `docs/plans/` are point-in-time planning artifacts. They intentionally reference specific line numbers, function signatures, and code state as it existed when the plan was written. Do not suggest removing line numbers or updating wording to stay current -- these documents are not maintained after implementation.
- **Scrut tests use intentional edge-case arguments**: Tests in `tests/*.md` sometimes pass values like color names (e.g., `orange`) as email ID arguments to boolean flags. This is intentional edge-case testing that verifies boolean flags do not consume the next argument as a value. The `authentication_failed` error is expected since scrut tests run without credentials. Do not suggest changing these arguments or flagging the authentication error as incorrect.
- **Markdown tables use standard pipe format**: Tables in this project use standard pipe-delimited Markdown syntax with a single leading `|` per row. Do not flag correctly formatted tables as having extra or duplicate pipe characters.
