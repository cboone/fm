# Exact sender and domain filters

Add `--from-exact` and `--from-domain` to search and the existing action
filter flags. The [issue](https://github.com/cboone/fm/issues/26) remains open.
The [initial check](../reviews/2026-11-19-exact-sender-initial-check.md)
records the missing feature and protocol evidence.

TOOLING relayed Chris's decision in mailbox message 000001. Narrow candidates
with a server From text search, then compare their parsed From addresses in
the client. Use the existing case-insensitive address comparison and ASCII
domain comparison. Apply every supplied predicate. Both exact predicates, if
supplied, must match the same address. Existing filters still apply.

1. Share sender query and query-state checks with the existing sender lookup.
   Preserve server sort order and apply the search offset, limit and total to
   confirmed matches. Action queries receive only confirmed IDs.
1. Cover false positives, combinations, page boundaries and query changes
   with synthetic JMAP fixtures. Update help and document cleanup previews.
1. Run the full local review order and validate each round against its scope
   captures. Run plants and final repository checks, tidy the signed commits
   and open the ready PR.

Work started from updated main at `bc9962f`. The checkout uses Go modules and
has no npm manifest or lockfile. Use the repository checks and installed tools.
All mail and JMAP responses used in validation are synthetic.
