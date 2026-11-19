# Bare unsubscribe values

Issue: [bare unsubscribe values](https://github.com/cboone/fm/issues/87).
Base: `origin/main`, `ba9ad31d`. Profile: `normal-max-default`.

## First check

The current inspection parser ignores bare HTTP, HTTPS and mailto values while
reporting the POST marker independently. Synthetic reproduction is retained at
`worker/evidence/first-check.log` in this workstream's mailbox. The separate
execution parser added by PR 93 does not change inspection. The issue remains
worth fixing because inspection hides usable unsubscribe mechanisms. No earlier
unmerged fix or overlapping open PR was found. TOOLING confirmed exclusive
parser ownership in mailbox reply 000001.

## Implementation

1. Extend inspection parsing to accept one bare URI when no angle brackets are
   present. Preserve bracketed alternatives and commas within URI payloads.
1. Report one-click only with a parsed HTTPS URL and the POST marker. Preserve
   raw-header reporting. Keep execution eligibility under its existing stricter
   parser; discovery does not authorize a request.
1. Cover bracketed, bare, malformed, MIME-decoded and unsupported values with
   synthetic unit cases. No mail, credentials, account or endpoint access.
1. Run the full structural and focused review order, trims, plants, final lint
   and contained CLI tests; tidy into signed commits, then open a ready PR.
   Never merge.

## Setup and evidence

This Go repository has no `package.json`, npm lockfile or local install script,
so `npm ci` does not apply. Use existing Go tooling and the repository-pinned
Scrut release with checksum verification in a worktree-local tools directory.
All CLI tests must use `tests/run-cli-tests`. Store raw checks and review output
outside the checkout in the mailbox; commit the plan, ledger and plant fixtures.
