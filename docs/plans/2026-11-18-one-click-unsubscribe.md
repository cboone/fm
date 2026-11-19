# Reviewed sender-list unsubscribe

Implement the authorized cut of [the unsubscribe feature](https://github.com/cboone/fm/issues/88).
Keep existing inspection and mailto draft modes. Add `--one-click` with repeatable
`--sender` addresses, preview by default and `--confirm` for a single POST per
eligible sender. Reject mixed modes before authentication.

Select the newest exact From address, paging past fuzzy matches. Fetch only
metadata and ordered headers. Require unambiguous unsubscribe headers, one HTTPS
target, and a sender-domain DKIM pass in Fastmail's receiving authentication
results. Unknown provenance or authentication means skip. Keep URL capabilities
out of output. Use a separate timeout-bound HTTP client with no credentials,
cookies, proxies, redirects or retries, and validate resolved dial addresses.

Reuse existing JMAP query/get types and formatter conventions. The inspection
parser and authenticated retrying transport are unsuitable for execution.
Fastmail's sender-authentication and delivery-process documentation establish that
inbound authentication results are provider-added. The implementation must
distinguish those from sender headers.

Tests use synthetic JMAP data, transports without external dialing, or explicitly
scoped local servers. No real mailbox or unsubscribe endpoint is contacted.
Cover selection, authentication, parser ambiguity, address policy, preview,
isolated POST behavior, partial failures and redaction. Plant guard defects with
the installed checked runner. Run structural/cleanup review, focused review,
trims, lint, the repository's final checks and exact-tree tidy. Persist rounds
and findings in a review ledger. Stop before push and report to TOOLING.

Record discovery, combined reporting, thresholds and file input as deferred on
the issue. The explicit list selects one message's list per sender, not every
subscription associated with that address.
