# Bare unsubscribe review ledger

Base: `origin/main` at `ba9ad31d`. Scope: bare-value inspection and the approved ASCII protocol-comparison class, issue 87.
Worker profile: `normal-max-default`. TOOLING holds parser serialization for
this workstream, acknowledged in mailbox reply 000001. No mail, credentials,
account access or external unsubscribe requests. CLI checks use only the
contained wrapper. No merge is authorized.

Evidence paths below are relative to
`/Users/ctm/.local/state/codex-mailboxes/tooling/87-bare-unsubscribe-urls/worker/evidence/`.
The complete installed plant-defects 1.3.0 bundle is snapshotted unchanged at
`plant-helper/`; `plant-helper.json` records every file hash and mode.

## Structural loop

Round 1, Nit threshold: independent standalone read-only Codex executions,
`gpt-6.1-sol`, high effort, 872000 context window, web disabled. Anti-patterns
and all four cleanup angles completed; result-shape checks exited 0 in the
strict call recording this section. Snapshot and recapture in `structural1/`
found no lasting changes. No editing simplify ran.

| Finding                    | Reviewer label | Our severity                                      | Take or record                                      | Review again            |
| -------------------------- | -------------- | ------------------------------------------------- | --------------------------------------------------- | ----------------------- |
| Partial POST markers       | important      | Important: inspection advertises false support    | Take: exact case-insensitive trimmed marker         | Yes, changes inspection |
| URL judge unrelated error  | important      | Important: plant evidence accepted a weaker judge | Take: isolated wrong-URL canary and judge weakening | No, tests only          |
| Full URL lowercase         | suggestion     | Nit: avoid transforming the entire capability URL | Take: retain scheme classification on selection     | No, equivalent result   |
| Name-based mailto expected | suggestion     | Nit: scattered test expectations                  | Take: put subject and body in the table             | No, tests only          |

Round 1 corrections: `5534e2d`. The URL judge now supplies only a wrong URL
expectation, with correct mechanism and indicator expectations. The check
requires that canary to fail; removing its URL comparison must make the check
reject the judge. Initial plant evidence is superseded. The next structural
round runs at Important because marker acceptance changed behavior.

## Claims and instruments

| Changed behavior                      | Synthetic tests                         | Plant                           |
| ------------------------------------- | --------------------------------------- | ------------------------------- |
| Single bare HTTPS/HTTP/mailto         | `TestParseBareAndOneClick` scheme cases | bare-fallback/http/mailto       |
| Reject whitespace and missing host    | malformed bare and bracketed cases      | whitespace/host                 |
| Indicator requires HTTPS and marker   | HTTP/mailto/empty/missing-marker cases  | oneclick-target/marker          |
| Marker must be complete               | prefixed and suffixed marker cases      | complete-marker                 |
| URL assertion distinguishes wrong URL | controlled wrong-URL canary             | judge-url                       |
| Bracketed alternatives, comma payload | existing and expanded parse tables      | covered by fallback restoration |

Raw headers, mailto parameters and encoded-word compatibility have regression
assertions; their underlying decoding behavior is unchanged. Inspection
`one_click` describes advertised HTTPS support, not authenticated execution
eligibility. The separate execution parser remains strict and unchanged.

## Structural round 2 stop

Round 2, Important threshold: all five standalone reviewers completed with
validated outputs. Result validators exited 0 in the strict call recording this
section. Scope recapture found no lasting changes. Evidence: `structural2/`.
End state: **stopped: recurring defect**. No correction is authorized until
TOOLING approves the mechanism audit. Counts remain structural 2, focused 0,
comment trims 0, prose trims 0, Copilot 0.

| Finding                       | Reviewer label | Our severity                                 | Take or record                            | Review again            |
| ----------------------------- | -------------- | -------------------------------------------- | ----------------------------------------- | ----------------------- |
| Unicode lookalike POST marker | important      | Important: invalid marker advertises support | Take after marker-matching class approval | Yes, advertised support |
| Repeated bare-header scan     | suggestion     | Nit: redundant traversal                     | Record below Important                    | No edits                |
| Positional table literals     | suggestion     | Nit: readability of adjacent string fields   | Record below Important                    | No edits                |

`Liſt-Unsubscribe=One-Click`, with U+017F replacing ASCII s, is accepted by
`strings.EqualFold`. The execution parser rejects it. This repeats the invalid
POST-marker acceptance mechanism corrected in round 1. Proposed class audit:
complete ASCII protocol-marker matching across inspection, execution and their
fixtures, including every non-ASCII rune that Unicode folding treats as an
ASCII marker letter. Preserve the existing deliberate ASCII case-insensitive
inspection contract and the stricter execution contract. No endpoint request is
needed. TOOLING must approve this class before corrections or another round.

Later focused review, trims, final lint, final receipts, tidy, push and PR
creation have not run. Preliminary test and plant output is retained as
superseded evidence, not a final check or ready-PR claim.

## Approved comparison-class audit

TOOLING replies 000002 and 000003 approved the class and identity extension:
protocol-marker comparison uses Unicode case folding, so non-ASCII runes match
ASCII markers. Provider-host, authserv-id and sender domains fail closed on
non-ASCII. Preserve local-part handling, valid ASCII alignment and strict
execution-marker spelling. Reply 000002 explicitly said no restart was needed.
Correction: `04dd609` and its follow-up. Structural count stays 2; the next
round remains Important. All prior plant results are superseded.

| Site checked                                         | Disposition and instrument                                                                                  |
| ---------------------------------------------------- | ----------------------------------------------------------------------------------------------------------- |
| `Parse` scheme switch and HTTPS selection            | URL parser requires ASCII scheme; shared ASCII comparison; parse cases                                      |
| `isOneClick` marker                                  | Shared ASCII comparison; marker ASCII/Unicode cases; marker-wiring plant                                    |
| `EqualFoldASCII` inputs                              | Both operands require ASCII; exhaustive SimpleFold orbit tests, explicit long s/Kelvin; guard plant         |
| `EqualSenderAddress` domain                          | ASCII comparison after final @; local-part folding preserved; helper and domain plant                       |
| `Target` two header names                            | ASCII comparison; protocol tests; separate link/post wiring plants                                          |
| `Target` POST marker and URI scheme                  | Exact comparisons unchanged; canonical/lowercase/Unicode cases                                              |
| `Authenticated` Received header                      | ASCII comparison; existing provider-boundary fixtures                                                       |
| Internal host suffix                                 | ASCII guard before lowercasing; dotted-I fixture and plant                                                  |
| Provider-host regex                                  | ASCII guard before case-insensitive regex; long-s/Kelvin provider fixtures and plant                        |
| Authserv-id comparison                               | ASCII comparison; lookalike result fixture and plant                                                        |
| Authentication-Results header                        | ASCII comparison; long-s fixture and plant                                                                  |
| DKIM-Signature header                                | ASCII comparison; Kelvin/long-s fixtures and plant                                                          |
| DKIM pass token                                      | ASCII comparison; Kelvin/long-s fixtures and plant                                                          |
| Result and signature sender domains                  | ASCII comparison; independent lookalike fixtures and plants                                                 |
| Authentication property keys                         | Lowercase map has only `header.d`/`header.b` consumers; no non-ASCII lowercase maps to their letters        |
| DKIM signature tag keys                              | Exact `d`/`b`/`h` keys unchanged                                                                            |
| Signed `h` header names                              | ASCII guard before lowercasing; long-s/dotted-I fixtures and plant                                          |
| Received/LMTP grammar regexes                        | Fixed literals have no Unicode folding equivalent; provider and internal host identities guarded separately |
| Client newsletter List-Id and List-Unsubscribe names | ASCII comparison; client name cases; independent wiring plants                                              |
| Client inspection unsubscribe and POST names         | ASCII comparison; client name cases; independent wiring plants                                              |
| Client newest-sender From match                      | Shared address comparison; synthetic query/get fixture and plant                                            |
| Command selected From match                          | Shared address comparison; preview-only synthetic fixture and plant                                         |
| Command sender validation and deduplication          | Non-ASCII domain refused before existing lowercase dedup; sender cases and validation plant                 |
| Request body marker                                  | Fixed canonical constant, unchanged; no header-derived spelling                                             |
| Generic mail summary domain/grouping                 | Outside unsubscribe eligibility; unchanged                                                                  |

The dotted capital I U+0130 also maps to ASCII i during `ToLower`, though not
`SimpleFold`; it is covered in marker, header-name, signed-coverage and internal
host cases. `SimpleFold` orbits cover every non-ASCII rune equivalent to an ASCII
letter in the installed Unicode tables. Fixtures contain explicit U+017F and
U+212A cases. None of these checks needs a mailbox or unsubscribe request.

## Structural loop completion

Round 3, Important threshold: all five reviewers completed and the recapture
found no lasting changes. One Important finding confirmed the plant test
selection omission also reported by plant run 4. No additional Important
production comparison defect was found. The correction is a test-filter edit
and an equivalent if/else avoiding redundant bracket extraction.
Targeted Go checks exited 0 in the strict call recording this section.

| Finding                         | Reviewer label | Our severity                                    | Take or record                                                                      | Review again            |
| ------------------------------- | -------------- | ----------------------------------------------- | ----------------------------------------------------------------------------------- | ----------------------- |
| Sender-domain test not selected | important      | Important: required plant evidence cannot pass  | Take: select `TestEqualSenderAddressASCII`                                          | No, test selection only |
| Duplicate header scan           | suggestion     | Nit: avoid an unused extraction result          | Take: select bare or bracketed extraction explicitly                                | No, equivalent results  |
| Input address self-comparison   | suggestion     | Nit: explicit domain predicate could be clearer | Record: shared comparison keeps rejection policy in one place; extra API adds scope | No edits                |

Structural end state: **ended: fixes changed no behavior**. Count 3 (Nit,
Important, Important). Focused review follows. Prior positional-test suggestion
remains recorded below threshold. No editing simplify ran.

Plant run 3 failed with inputs changed during the ongoing audit. Plant run 4
stopped at the sender-domain variant because its named test was excluded;
later variants were unrun. Neither run is complete passing evidence. The
corrected filter will be exercised in a fresh run and in the final receipt.

## Focused loop completion

Round 1, Nit threshold: code, errors, tests, comments and types all completed
with no findings, validated output and recorded process exit 0. Status validation
and parent repository-metadata refresh exited 0 in the strict call recording
this section. The error reviewer could not reach GitHub metadata from its
sandbox; the parent independently confirmed the repository remains active and
not a fork. Scope recapture found no lasting changes. Evidence: `focused1/`.
All five aspects applied because behavior, error boundaries, tests, documentation
and output invariants changed. No editing simplify ran.

Focused end state: **clean at Nit**, count 1. Structural count 3, ended: fixes
changed no behavior. Copilot count 0. Comment and prose trims follow once each.
Plant run 5 is preliminary output. The final gate will run and render the
plants with a tree-bound receipt.

## Trims and final validation

Comment and prose trims each ran once through standalone read-only Codex,
`gpt-6.1-sol`, low effort, 872000 context window. Both processes exited 0 with
complete nonempty reports under `trims/`. Comment trim proposed no changes;
unchanged-code comparison was unnecessary. Prose trim took both wording
proposals, in the CLI reference and plan. No proposals were declined or flagged;
historical evidence, gates and scope constraints were retained.

The final check receipt is `final-gate/receipt.json` under the evidence root.
Its tree verification determines the suite, lint and plant status, rather than
this ledger. Final plant status and generated ledger are outside the checkout
at `final-plants-status.md` and `final-plants-ledger.md`. The root plant manifest
and all fixtures remain committed for reproduction.

`lint-and-fix --no-push` runs required `make fmt` and `make lint`. ShellCheck,
shfmt and formatting/lint of the new Markdown records are additional scoped
checks. The existing CLI reference is edited without whole-file reformatting.
Final gate, signed exact-tree tidy and PR creation remain pending in this record;
the handoff supplies their receipts and final commit.
