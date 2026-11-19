# Exact sender filters initial check

Stopped before implementation for a selection-strategy decision. The feature
remains missing and worthwhile: exact address/domain selection would reduce
accidental selection during triage, including existing action-query mode.

## Repository and issue evidence

- Checked on 2026-11-19 at 03:12 AM EST. `origin` is `cboone/fm`, with no upstream
  remote. GitHub reports `isArchived: false`, `isFork: false`, `parent: null`.
- Fetched `origin/main` and worktree HEAD are both
  `9110c04bd9b623c025fe4cc70d3dc4a20feefbd9`. The worktree began clean.
- [Exact sender/domain issue](https://github.com/cboone/fm/issues/26) is open,
  marked `in progress`, with no comments. It requires deterministic,
  case-insensitive matching and combinations with existing filters.
- The open PR list contains only the gh-actions pin update, PR 85. Searches of
  all PRs and advertised remote/local branches found no earlier exact-filter
  implementation to adopt.
- `cmd/search.go` and `cmd/filters.go` expose only text-based `--from`.
  The existing Sieve `--from-domain` creates incoming rules; it does not select
  stored messages for search or actions.

## Contract and prior art

`internal/client/email.go:347` builds the server filter; `SearchEmails` applies
offset, limit and sorting on `Email/query`, and returns its total.
`QueryEmailIDs` collects every server-selected ID. `QueryFirstEmailID` requests
only the newest server match. These contracts need exact selection before
pagination or action resolution, rather than filtering a returned page.

[RFC 8621 section 4.4.1](https://www.rfc-editor.org/rfc/rfc8621.html#section-4.4.1)
defines `from` and `header` as text searches. It leaves string-match semantics
to implementations and does not provide parsed-address equality. Quoting
requests a phrase, not equality of the parsed address or its domain.

[Fastmail search help](https://www.fastmail.help/hc/en-us/articles/360060591213-Searching-your-mail)
documents address and `@domain` searches in its UI. That page does not establish
an equality contract for the JMAP `from` field.

Existing `NewestSenderEmail` in `internal/client/unsubscribe.go:99` pages fuzzy
sender candidates, checks parsed From addresses locally, and aborts on changed
query state. `EqualSenderAddress` provides a reusable case-insensitive address
comparison. Extend those mechanisms where appropriate rather than adding a
separate tool. Neither defines complete exact-search totals or pagination.

## Decision required

Recommend a client-side scan of messages matching the other existing filters,
checking parsed From addresses before counting and paginating exact matches.
This avoids assuming the fuzzy server filter returns every case-insensitive
exact match, but broad searches can require reading metadata for the entire
candidate set. Action-query mode needs the same exact selector.

TOOLING must settle whether that scan/cost is in scope, or whether a documented
Fastmail-specific candidate-narrowing contract should be established first.
Do not silently return server totals for locally filtered results or cap the
candidate scan and present partial matches as complete.

No production code changed. No live mailbox, credentials, message actions,
account changes, installs, background jobs, releases, tags or PRs were used.
This checkout has no npm manifest/lockfile or repository tool installer;
dependency setup, tests, plants and local review rounds have not started.
