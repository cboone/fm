# Draft error details review

Scope: `fix/91-draft-error-details`, base `origin/main`, merge base
`34b5ddc1d2613e264dfbccb3896c90a800594ec9`. TOOLING approved the scope and
restart in reply 000001: proceed independently of the built one-click branch,
use synthetic unit tests, skip CLI tests requiring its containment harness, and
merge main before the final gate if the one-click branch lands first.

## Reviews

Structural round 1 ended **clean at Nit**. Anti-patterns and all four cleanup
reviewers completed with no findings. Anti-pattern JSON passed the installed
validator. Every cleanup run exited 0 and validated. Scope comparison found no
lasting changes, tree `b5654fbeeed2900b626da0caae0a453d364ae634`.

Focused round 1 ended **clean at Nit**. Code, errors, tests and comments each
completed with exit 0 and validated output, with no findings. No type
definitions changed, so types did not apply. Editing simplify was not selected.
Scope comparison found no lasting changes, tree
`a7c30caa5adf2f917c5ac65a7e588bc8347a01bb`.

Both loops ran as separate read-only Codex executions, `gpt-6.1-sol`, high
effort, 872000 context window, web disabled and approvals never. The sandbox
could not reach GitHub for metadata; the worker independently verified current
remote ownership and non-archived/non-fork status. Structural cleanup used the
default AGENTS guidance; focused review also passed the configured CLAUDE
fallback name.

Comment and prose trims each ran once, separately, read-only, `gpt-6.1-sol`, low
effort, 872000 context window. Comments proposed no changes. The prose
suggestion split one scope paragraph into two requirements and removed
“actually”; taken, with meaning and approval boundaries preserved. No proposals
were declined.

Counts: structural 1; focused 1; comment trim 1; prose trim 1; Copilot 0. There
were no findings requiring severity, take/record or recurrence decisions. Only
the prose trim changed the reviewed tree afterward. Plant inputs were added
after structural review and included in the focused review.

## Coverage and final gate

Synthetic SetError cases cover type-only, type plus description,
description-only, empty description, empty error and null error. Synthetic
identity cases cover missing and present submission; the missing case refuses
any transport invocation. Ordinary draft success without submission guards
against requiring sending permission. Existing tests cover cached identities,
transport failures and draft modes. All new behavior uses the existing injected
client seam; no credential helpers or network transports are involved.

The generated plant results below verify the error type, description, capability
guard, scope advice and absence of a general submission prerequisite.

Final checks are authoritative only through the checked-gate receipt at
`worker/evidence/final-gate/receipt.json`, verified against the final tree. The
gate includes `make fmt`, `make vet`, `make lint`, `make test`, shell checks,
Markdown checks and current-input plant verification. `lint-and-fix --no-push`
uses the repository's enforced Go checks; shell checks are advisory, and
Markdown formatting and lint follow the loaded writing skill. No install is
needed. The dependent action rebuilds three logical signed commits and verifies
tree equality with the pre-tidy backup. The receipt records each exit and the
action.

CLI and live tests are intentionally unrun under TOOLING's synthetic-only scope.
No account, credentials, mail, real drafts, live endpoints, installation, live
configuration, release or tag was used. No merge is authorized.

Evidence paths above are relative to
`/Users/ctm/.local/state/codex-mailboxes/tooling/91-draft-error-details/`.
Review output is under `worker/evidence/reviews/structural1/` and `focused1/`;
trim prompts, outputs and logs are under `worker/evidence/`.

<!-- checked-plants:start -->

## Verified plants and controls

plant drop-type: named assertion reached; red exit 1; restored; green exit 0.

plant drop-description: named assertion reached; red exit 1; restored; green
exit 0.

plant skip-capability-check: named assertion reached; red exit 1; restored;
green exit 0.

plant drop-scope-advice: named assertion reached; red exit 1; restored; green
exit 0.

plant require-submission-for-drafts: named assertion reached; red exit 1;
restored; green exit 0.

Evidence:
`/Users/ctm/.local/state/codex-mailboxes/tooling/91-draft-error-details/worker/evidence/plants1/results.json`.

<!-- checked-plants:end -->
