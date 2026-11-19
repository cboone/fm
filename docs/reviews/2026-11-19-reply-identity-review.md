# Reply identity review

Scope: `fix/89-reply-identity`, base `origin/main` at `9110c04`.
Policy and setup decisions are recorded in the plan and TOOLING mailbox replies
000001 and 000002. No merge or live access is authorized.

## Review ledger

Structural round 1: anti-patterns and all four cleanup reviewers completed
with exit 0 and validated output. The strict wrapper captured scope before and
after every reviewer stopped. Comparison exited 0, `scope_changed` was false,
and `changed` and `outside_scope` were empty. Evidence: `structural1/`.

- Reviewer label `nit`: CLI reference contract drift. Our severity: Nit,
  missing documentation for this change. Take: update sender rules and result
  schema. Review again: no, docs only.
- Reviewer label `suggestion`: recipient slice copying. Our severity: Nit,
  avoidable allocation during matching. Take: iterate To and Cc directly.
  Review again: no, selection semantics unchanged.
- Reviewer label `suggestion`: temporary boolean lookup map. Our severity: Nit,
  unnecessary test naming state. Take: format the boolean directly. Review
  again: no, tests only.

Structural end state: **ended: fixes changed no behavior**. Structural count 1;
focused count 0; comment/prose trims 0; Copilot 0. Reviews used separate read-only
Codex executions, `gpt-6.1-sol`, high effort, 872000 context window, web disabled
and approvals never. No editing simplify ran. The command supplied verified
repository metadata and the absence of a branch PR to the anti-pattern reviewer.

## Validation status

Final validation is pending. An initial plant run is incomplete: the warn-in-json
patch left an unused import and failed setup, exit 1. That fixture now retains
the import use; the complete plant set will run again. No successful aggregate
plant claim comes from the failed run. Evidence: `plants1/results.json/`.

The final gate will run the repository CI target, the synthetic plants and their
render, formatting/lint checks, and tree verification. Live tests are excluded.

Evidence paths are relative to
`/Users/ctm/.local/state/codex-mailboxes/tooling/89-reply-identity/worker/evidence/`.

## Focused round 1 stop

All five aspects completed with exit 0 and validated output. The strict wrapper
captured scope before and after every reviewer stopped; comparison exited 0,
`scope_changed` was false and both changed-path lists were empty. Evidence:
`focused1/`. Counts now: structural 1, focused 1, trims 0, Copilot 0.

- Reviewer label `suggestion`: JSON warning field-name assertion. Our severity:
  Nit, because the production contract is correct but its regression test
  round-trips through the same type and would accept a renamed key. Take:
  independently assert the literal warnings key and plant its rename.
  Review again: no, test-only change.
- Reviewer label `suggestion`: From omission note ignores automatic selection.
  Our severity: Nit, because a selected identity can supply From even with a
  non-email session username. Take: correct the schema note. Review again:
  no for this edit, docs only; the recurrence stop applies first.

End state: **stopped: recurring defect**. The structural round corrected sender
and warning documentation but left another old sender rule in the schema note.
Proposed audit mechanism: sender selection and fallback contract duplicated
across command help, README, CLI reference, bundled runbook and result schema.
Await TOOLING approval for that class audit before further fixes or reviews.
No production behavior or policy expansion is proposed. At this stop, the branch
was unpushed with no PR; trims, final gate and tidy were unrun.

## Approved contract audit

TOOLING reply 000003 approved the class: duplicated sender selection and warning
contracts drift from code. The audit and fixes are tests
and docs only, with no policy change; no further review round is required under
Findings. The recurrence stop is resolved for this branch. Focused end state:
**ended: fixes changed no behavior**. Counts remain structural 1, focused 1,
trims 0, Copilot 0.

Sites checked against implementation:

- `cmd/draft.go` long help and From flag: To then Cc, first match, override,
  default retention, missing-capability warning and output streams agree.
- `cmd/draft.go` result formatting: text warnings use stderr; JSON serializes
  the optional array through the existing formatter. No new formatter rule.
- README setup scope: clarified basic read/triage versus draft identity lookup.
  Drafting Protocol: recipient order, case matching, sender name, override,
  fallback, other errors and both warning representations agree.
- CLI reference draft flag table and From identity paragraph: configured
  identity override, first matching recipient and existing default agree.
- CLI reference identity warnings: only missing submission falls back; explicit
  From and other errors retain failures. JSON key and text stream agree.
- CLI reference reply/reply-all/forward bullets: recipient and threading rules
  remain true. JSON/text examples remain valid no-warning/default-sender cases.
- CLI reference DraftResult: corrected the From omission note to cover selected
  identities; From and warnings field rows agree with construction/encoding.
- Bundled runbook Draft Composition: recipient order, override, fallback,
  missing capability, other errors and warning inspection agree.
- `internal/types/types.go` DraftResult: From and warnings retain optional JSON
  tags; the warning test now asserts the public key independently of this type.
- `internal/client/draft.go` options, sender resolution, selection, warning,
  reply-all exclusion and result construction: comments and output agree.
- `internal/client/identity.go` capability error and identity resolution:
  explicit lookup keeps the merged scope advice unchanged.
- Help snapshot: the added sender and warning text matches command help.
- Plan and this ledger: policy, scope, streams and authority remain consistent.
  Plant patches intentionally model defects; they are not product guidance.

Claim-to-test mapping:

- Recipient order, case-insensitive address match, identity name, no match,
  explicit override and reply-all exclusion: `TestCreateDraft_ReplyIdentity`.
- Missing-capability fallback, retained default, warning and explicit error:
  `TestCreateDraft_ReplyIdentityMissingCapability`.
- Other lookup failures remain errors: `TestCreateDraft_ReplyIdentityLookupError`.
- New/forward sender rules: `TestCreateDraft_NonReplyNoIdentityLookup`.
- Literal JSON warnings array, omission and text stderr: `TestDraftWarningsOutput`.
- Each weakened selection/output boundary has a fixture in the plant manifest;
  rename-warnings-key and keep-empty-warnings cover the strengthened assertions.

## Trims and final validation contract

Comment and prose trims each completed once, exit 0, in separate read-only
Codex runs using `gpt-6.1-sol`, low effort and an 872000 context window.
The strict wrapper captured scope before and after both runs stopped;
comparison exited 0 with `scope_changed` false and both path lists empty.
Evidence: `trims1/`. The comment rewrite and all seven prose proposals were
taken. None were declined; no behavior, policy, counts or evidence changed.
Only the intended comment bytes changed in source; affected patch context was
updated. Counts: structural 1, focused 1, comment trim 1, prose trim 1,
Copilot 0.

Both review loops end **ended: fixes changed no behavior**. The approved
recurrence audit is resolved. No additional correctness round is required for
the subsequent tests, docs and comment edits under the conductor decision.

`lint-and-fix --no-push` uses scoped gofmt, golangci-lint auto-fix and the
repository Go checks; fixture shell and new process-document checks accompany
it. Existing README/CLI-reference formatting and lint issues are advisory,
not a claim that their entire files are clean. The Markdown checks were run,
and baseline copies retain the findings; unrelated reformatting was excluded.

Final claims are authoritative only through `final-gate/receipt.json`, verified
against the tidied commit tree. That gate runs `make test-ci` (vet, fmt, lint,
unit and contained CLI tests), shell checks, formatting/lint on the new plan
and review record, whitespace checks excluding unified patch artifacts, and a
fresh plant run plus render. Generated plant statuses and ledger are external
artifacts `final-plants-status.txt` and `final-plants-ledger.md`, never copied
into a success claim by hand. The owned plant-defects 1.3.0 helper snapshot and
hash/mode manifest are `plant-defects/` and `plant-defects.json`.

Tidy backs up the unpushed branch and reconstructs logical signed commits from
its original base, then requires byte-for-byte Git tree equality. No live
account, mail, credentials, draft, configuration, install, release or tag was
used. Live tests remain unrun. No merge is authorized.
