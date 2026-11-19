# Reply recipient identity

Scope: fix the remaining reply sender defect in
[the reply identity issue](https://github.com/cboone/fm/issues/89).
Current main and this branch started at `9110c04`; no covering branch or PR
exists. The merged draft error handling remains authoritative for explicit
`--from` and draft server errors.

TOOLING settled setup in mailbox reply 000001: use installed tools and repository
commands, with no npm install or manifest when no lockfile exists. Reply 000002
settled recipient priority and warning output and authorized continuing through
the ready PR. No merge, release, live configuration, account, credentials or
real mail access is authorized.

1. Extend existing draft creation and identity lookup. Replies without `--from`
   match To, then Cc, preserving recipient order and ignoring address case.
   Identity address/name become From. No match retains the existing default.
1. Missing submission capability alone falls back with a warning. Explicit
   `--from` and other lookup failures retain error handling. JSON carries an
   optional `warnings` array; text writes `Warning:` lines to stderr.
1. Cover both reply modes, recipient priority, override, fallback, errors,
   reply-all exclusion and output with injected synthetic JMAP responses.
   Update help, snapshots, README and the bundled runbook.
1. Run the full local review order with scope captures and an explicitly checked
   unchanged-scope comparison for every round. Plant changed behavior, run
   lint-and-fix without pushing, tidy signed commits and verify tree equality.
1. Produce a final checked-gate receipt, push and open a ready PR. Report to
   TOOLING and stop before merging.

Existing identity lookup, client injection and draft result formatting
suffice; no new tool or dependency is needed. Installed tools: Go 1.27.3,
golangci-lint 2.15.0, Scrut 1772332969, Prettier 3.8.1 and markdownlint-cli2
0.21.0. The CI Scrut install is checksum-pinned for Linux; no global install
was used.
