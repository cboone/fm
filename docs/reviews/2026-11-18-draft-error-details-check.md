# Draft error details initial check

The error-detail defect remains real and worth fixing. TOOLING answered the
scope and serialization questions in mailbox reply 000001 and lifted the
restart gate: build on main with synthetic unit tests, skip CLI tests requiring
the one-click branch's containment, and proceed through local review to a ready
PR. If the one-click branch merges first, merge main before the final gate.

## Evidence

- Fetched `origin/main` is `34b5ddc1d2613e264dfbccb3896c90a800594ec9`.
  `cboone/fm` is neither archived nor a fork. The worktree began clean.
- Read issues 91, 89 and 88, including comments, and the current PR list.
  No open PR covers this defect.
- Built branch `feat/88-one-click-unsubscribe` at `608f0c2` leaves
  `internal/client/draft.go` and `internal/client/identity.go` unchanged.
  It changes the shared client and unsubscribe command and contains the
  credential-refusing CLI test wrapper absent from this base.
- Synthetic JSON SetError responses through the existing injected client seam
  reproduce `forbidden` becoming `draft creation failed: unknown error` and
  `invalidProperties` with description becoming only
  `draft creation failed: synthetic rejection`.
- Installed go-jmap method declarations require only mail for `Email/set`,
  but submission for `Identity/get`. The unsubscribe draft path supplies no
  explicit From identity. Its rejection does not establish missing submission.

The reproduction check exited 0. Raw output and the reproducer are preserved at
`/Users/ctm/.local/state/codex-mailboxes/tooling/91-draft-error-details/worker/evidence/`:
`reproduction.log` and `issue91_repro_test.go`. The temporary test was removed
from the worktree after copying. No credential helper, account, draft, mail or
live endpoint was used. No production code changed. Full tests and local review
were not run because implementation has not started.

## Approved scope

Proceed independently of the built one-click branch on main. Preserve SetError
type and description for every draft rejection. Name submission plus the Email
submission token scope only when identity lookup requires that missing
capability. Do not imply
that mail-only draft creation requires sending permission or explain an unknown
server rejection as a missing submission capability.

No background gate is running. No merge, release, tag or installation is
authorized. Review and validation results follow in the review ledger.
