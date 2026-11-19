# One-click unsubscribe review ledger

Scope: branch `feat/88-one-click-unsubscribe`, merge base
`34b5ddc1d2613e264dfbccb3896c90a800594ec9` with `origin/main`. Stop before push.
All reviews use separate read-only Codex executions, `gpt-6.1-sol`, high effort,
872000 context window, no web or repository execution. Real mailbox and
unsubscribe endpoint access is prohibited. See the containment evidence limit
below for the initial inherited CLI cases.

## Design decisions

CONDUCTOR, answering for TOOLING in mailbox replies 000001 and 000002, approved
the conservative provider-prefix authentication policy and DKIM signature-header
coverage check. Trust only the Fastmail receiving block above the SMTP boundary,
with exact sender-domain DKIM pass, matching signature domain and signature prefix
when provided, and both unsubscribe fields in `h=`. Skip unknown results.
Imported/rewritten messages cannot be distinguished from delivery by headers
alone. Document this limitation for the final PR decision; do not silently claim
cryptographic re-verification. The README and help record the assumption.

Fastmail documents provider-added authentication results in its
[delivery process](https://www.fastmail.help/hc/en-us/articles/1500000278262-The-email-delivery-process)
and [sender authentication](https://www.fastmail.help/hc/en-us/articles/1500000280461-Sender-authentication).
[RFC 8058](https://www.rfc-editor.org/rfc/rfc8058.html) specifies the POST and
signature-coverage requirements. These sources do not prove imported-message
provenance. Provider layout compatibility is tested with synthetic fixtures;
live-account compatibility remains unverified by design.

Deferred discovery, combined reporting, thresholds and file input are recorded
in [the issue comment](https://github.com/cboone/fm/issues/88#issuecomment-6995669039).

## Structural loop

Round 1: Nit threshold. Anti-pattern review and four cleanup reviewers completed
successfully. Reviewed snapshot `92fb81bd003457a5843adeeefe64667565f35041`;
scope recapture found no lasting changes. Artifacts:
`worker/evidence/reviews/anti.json` and `cleanup/*.json`.
Anti-pattern result passed its installed JSON validator. Each cleanup status was
`ok`, process exit 0. The anti-pattern sandbox could not read GitHub metadata;
the parent independently verified non-archived, non-fork `cboone/fm` before work.

| Finding | Reviewer label | Our severity | Take or record | Review again |
| --- | --- | --- | --- | --- |
| Comma inside bracketed HTTPS URI rejected | important | Important: eligible senders are silently skipped | Take: parse separators outside brackets; test raw/escaped commas, alternatives and malformed separators | Yes, target selection changes |
| Lookup error masks HTTP-only batch failure | nit | Nit: missing boundary test | Take: separate HTTP-only and success cases; add failure-flag plant | No, tests only |
| Unsubscribe schema omits mode distinction | nit | Nit: output documentation drift | Take: qualify inspection/draft schema | No, docs only |
| Duplicate raw header representation | suggestion | Nit: duplicate maintenance | Take: alias existing `types.Header` | No, representation unchanged |
| Timestamp conversion repeats `safeTime` | suggestion | Nit: duplicate maintenance | Take: use existing helper | No, equivalent conversion |
| Authentication parsed after invalid target | suggestion | Nit: discarded computation | Take with per-sender helper: validate target first | No, equivalent outputs |
| CIDR constants reparsed per address | suggestion | Nit: unnecessary parsing | Take: preparse constants | No, equivalent policy |
| Nested per-sender result flow | suggestion | Nit: control-flow readability | Take: helper with early returns | No, equivalent outputs |
| Extract authentication-result parser | suggestion | Nit: optional structural change | Record: keep the conservative grammar local; another parser layer adds scope without fixing behavior | No change |

Round 1 fixes are in `2868f5f`. The URI grammar change requires round 2 at
Important.

Round 2: all five reviewers completed, processes exited 0 and outputs validated.
Snapshot `e91e6547cadff1b507fe82b8666b5fcec6d08b0b`; recapture found no lasting
changes. Anti-pattern review found no Important findings or recurrence.
Cleanup reported one `suggestion`: parse message-dependent DKIM signatures once
per check. Our severity: Nit, repeated allocation with multiple pass records.
Record below threshold; do not change behavior or start another round. This is
distinct from constant CIDR setup, which was moved out of request processing.
Review again: no changes. Artifacts: `worker/evidence/reviews/round2/`.
Structural end state: **clean at Important**.

Structural review completed after two rounds; focused review follows below.

## Focused loop

Round 1: Nit threshold, five aspects (`code`, `errors`, `tests`, `comments`,
`types`). All completed, process exit 0 and schema validation succeeded.
Recapture found no lasting changes. Artifacts:
`worker/evidence/reviews/focused1/`. No editing simplify ran.

| Finding | Reviewer label | Our severity | Take or record | Review again |
| --- | --- | --- | --- | --- |
| Only first public DNS address dialed | important | Important: reachable endpoints fail when the first address is unavailable | Take: try validated addresses under the request deadline before HTTP execution | Yes, connection behavior |
| Setup and lookup causes discarded | important | Important: failures provide no corrective action | Take: safe credential/session/account/connectivity/changed-selection categories and hints | Yes, error contract |
| Policy refusals described as unknown delivery | important | Important: false uncertainty hides blocked destinations | Take: typed policy, DNS and connection errors; safe certificate category | Yes, error contract |
| Production address-guard wiring untested | important | Important: unsafe wiring could escape isolated helper tests | Take: production transport test scoped to an owned TLS fixture and a wiring plant | No, test only |
| Fastmail endpoint restriction untested | important | Important: another provider could supply supposed Fastmail evidence | Take: flag/environment CLI cases with refusing credentials and a plant | No, test only |
| POST capability preservation untested | important | Important: path/query changes could target the wrong subscription | Take: exact outgoing escaped path/query assertions and two plants | No, test only |
| Published internal LMTP layout rejected; SMTP HELO could obscure the boundary | self-found | Important: delivery compatibility and authentication provenance | Take: allow only internal LMTP before results; stop at SMTP regardless of HELO; positive/negative fixtures and plants | Yes, eligibility behavior |

The last finding came from comparing Fastmail's published delivery example with
the parser, not live mail. It is a distinct receiving-trace mechanism from the
URI delimiter fix in structural round 1. The URI propagation finding is missing
integration coverage, not another observed parser failure.

CONDUCTOR approved the minimal test containment scope in reply 000004:
commit refusing `security`/`secret-tool` fixtures, verify their selection in a
wrapper, clear ambient fm settings, and route `make test-cli` through it. Live
tests stay separate. A plant removes the stub PATH prefix and expects refusal
before any CLI test. No broader wrapper follow-up remains necessary.

The first regeneration script failed on insertion-only patches, but an
unguarded shell sequence still committed the source changes. The correction
validates every replacement before writing any patch and uses a subsequent
signed commit. No review or final gate uses the incomplete patch set.

Focused correction check record: targeted Go tests, isolated one-click scrut
cases and round-1 status validation each exited 0 in the same strict tool call
that records this paragraph. Counts: structural 2, focused 1, trims 0, Copilot 0.
Focused round 2 advances to Important because fixes changed behavior.

Round 2: all five aspects completed, process exit 0, validated outputs, and
recapture found no lasting changes. Artifacts:
`worker/evidence/reviews/focused2/`. **Stopped: recurring defect**.
TOOLING was sent a stop with two proposed mechanism audits; no further fixes
are authorized until its class decision.

| Finding | Reviewer label | Our severity | Take or record | Review again |
| --- | --- | --- | --- | --- |
| Metadata fallback still loses typed JMAP operation/cause | important | Important: service failures get account-permission advice | Take after class approval: preserve typed method errors and allowlisted categories | Yes, error contract; recurrence |
| Actual command does not test consent forwarding | important | Important: passing true at the command boundary could send during preview | Take after class approval: contained command factories and actual-flag tests | Yes if wiring changes; recurrence of helper-only guard coverage |
| Text output assertions accept JSON fallback | suggestion | Nit: missing coverage of the new text contract | Take after class approval: exact text expectations | No, tests only |
| README overstates connection-timeout categorization | suggestion | Nit: documentation differs from the conservative unknown-outcome fallback | Take after class approval: qualify connection-deadline failures | No, docs only |

Requested audit mechanisms: sanitized error boundaries that erase typed causes,
and tests whose injection seams bypass production safety wiring. Scope is every
new boundary from Cobra consent through lookup, eligibility, client construction,
POST and output. Ledger every checked site and plant changed behavior. The
focused count remains 2; after approved behavioral fixes, its next round is the
second Important round (round 3). Counts are never reset.

Containment follow-up within the approved harness: three inherited date-flag
fixtures explicitly use `echo test` credentials with a localhost endpoint.
Replace those credential commands with refusing commands before the final CLI
gate, keeping their argument-validation expectations. This prevents a socket
attempt when a validation assertion regresses.

## Containment evidence limit

An initial arguments-only scrut run completed 19 cases before the inherited
OS-keychain fallback was noticed. Its authentication-failure output does not
prove zero credential-helper or mailbox-network attempts in the old search
cases. No unsubscribe confirmation was run, and no credentials or mail contents
were printed. This limitation was reported to TOOLING. Subsequent CLI checks
use the committed refusing stubs and isolated home/settings. Final evidence
must identify that environment, rather than retroactively asserting the initial
run was contained.

## Validation evidence

Final repository checks remain pending. Earlier targeted output and plant runs
are retained outside the checkout; final check records and generated plant
evidence will be recorded against current inputs before the final report.
The manifest now includes no-retry, timeout and ambient-proxy variants as well
as opaque-URL and HTTP-only batch-failure variants.

## Remaining work

Focused Important round 3, trims, lint, final repository
checks, current-input plant evidence, signed exact-tree tidy, then mailbox
handoff to TOOLING. No push or merge is authorized in this worker.


## Approved mechanism audits

CONDUCTOR approved both code mechanism audits in reply 000005 and the non-live
credential/endpoint audit in reply 000006. The code recurrence gate is resolved
by the following whole-class inventory and corrections. Focused counts stay at
2 completed; round 3 is the second Important round. A further recurrence stops
again. No review count is reset.

| Error site | Audit result | Regression evidence |
| --- | --- | --- |
| Mode/address/flag validation | Static input errors; no capability values echoed; unchanged | `TestOneClickSenders`, negative CLI fixtures |
| Credential setup | Explicit credential stage, typed cause kept, no helper output exposed | `TestOneClickMailboxSetupStages`, setup-stage plant |
| Session setup | Typed account, HTTP and connection causes kept; unknown protocol failures labeled unclassified | Setup-stage and safe-failure tests |
| Query transport failure | `SenderLookupError` preserves query operation and cause | Network-operation test and plant |
| Query method error | Original `jmap.MethodError` retained through unwrap | Typed-method test and query-cause plant |
| Query absent/state-changed response | Invalid response and changed selection remain distinct | Invalid-response and newest-selection tests |
| Get transport failure | Get operation and original transport cause retained | Network-operation test and get plant |
| Get method error | Original typed method error retained | Typed-method test and get-cause plant |
| Get missing response/message | Invalid response versus disappeared selection distinguished; null entries rejected | Invalid-response tests and null guard plant |
| Public metadata diagnostic | Operation allowlist plus method/status categories; server descriptions and URLs excluded | Safe-failure tests and category/payload plants |
| Target/authentication refusals | Existing fixed skip reasons; unknown authentication says unverified | Target/authentication and actual-command skip cases |
| POST errors | Policy/DNS/exhausted connection/certificate errors distinct; other delivery uncertainty retained | Safe-error/Post tests and policy plant |
| Response close | Known HTTP status retained, static close-error message; unchanged | Static audit; response body never read |
| Text result writes | One checked write per row replaces ignored intermediate write errors | Exact text/write-failure tests and plants |
| Formatter/partial exit | All rows emitted before partial-failure exit; format errors propagate | Actual-command HTTP-failure test and exit plant |

The JMAP library can return untyped response-decoding/HTTP errors. No unsafe
string parsing is used to invent their cause: those results explicitly say
unclassified transport/protocol failure and suggest checking service status and
a fresh preview, rather than incorrectly diagnosing account permissions.

| Safety wiring site | Audit result | Regression evidence |
| --- | --- | --- |
| Cobra mode and sender flags | Real command path tested; new StringArray reset corrected in test helper | Actual-command tests, reset plant |
| Duplicate/exact senders | Repeatable list deduplicates before lookup; exact metadata selection retained | Duplicate-confirm and query/get tests |
| Fastmail endpoint restriction | Flag and environment cases stop before credential resolution | Owned-endpoint CLI fixtures and endpoint plant |
| Mailbox construction | Contained factory seam; real constructor tested only with a refusing command or injected non-dialing transport | Setup-stage tests |
| Newest selection | Query sorting, pagination, response reordering and selection changes audited | Newest-sender tests and fuzzy/order plants |
| From ambiguity | Actual confirm command skips ambiguous sender data | Command ambiguity case and existing plant |
| Target validation | Actual confirm command skips unsafe target | Command unsafe-target case; target plants |
| Authentication | Actual confirm command skips unknown DKIM | Command unknown-authentication case; authentication plants |
| Consent forwarding | Actual default/dry-run flags send zero, confirm sends once | Command tests and consent plant |
| HTTP construction | Production guarded dial path exercised only against an owned TLS fixture | Production guard test and wiring plant |
| HTTP protocol/URI | Separate transport, exact opaque URI/form, no ambient auth/cookies, redirect or retry | POST isolation/error/timeout tests and plants |
| Results/exit | Real command decodes expected JSON metadata, statuses and partial error; text checked exactly | Command/text/HTTP-batch tests and plants |
| Test network boundaries | All non-live fixtures inventoried; owned session fixture and refusing proxy; secret-source/endpoint preflight wired before execution | Wrapper integration self-test and preflight/data plants |

## CLI validation process correction

A second inherited flags run revealed the same incomplete-inventory mechanism:
explicit fake-token commands could override the refusing credential default and
reach Fastmail's session endpoint. Those early flags tests attempted fake-token
session requests before containment. They did not use a real token or perform
an unsubscribe; zero mailbox-network attempts cannot be claimed for those runs.
The validation stop covers fm's CLI targets, wrapper and direct Scrut execution.

The responsible conductor approved preserving literal fake-token tests while
containing endpoints, and authorized lifting the stop after proof. Every command
in all six non-live fixture files is now checked. Removed environment unsets
that could drop the isolated endpoint/credential default. Literal endpoint
overrides now use the owned fixture variable. Positive credential-command and
precedence cases retain their original fake values. Empty, unset, keychain-capable
or unrecognized credential commands and external endpoint/proxy overrides are
rejected before Scrut starts. Live fixtures remain excluded and unchanged.

| Fixture | Credential/endpoint inventory |
| --- | --- |
| errors.md | Fake invalid-token case retained; endpoint inherited from owned fixture; no keychain fallback |
| flags.md | Fake environment/flag credentials retained; explicit localhost endpoints replaced by owned variable; no removed isolation settings |
| arguments.md | Refusing inherited credential default and owned endpoint |
| help.md | No authentication calls; inherited isolation retained |
| sieve.md | Refusing inherited credential default and owned endpoint |
| oneclick.md | Refusing credentials; endpoint rejection uses owned non-Fastmail endpoint; conflict validation remains before setup |

The owned fixture always returns HTTP 401 and never forwards requests. Its
refusing HTTP/HTTPS proxy also keeps an accidental default-endpoint fallback
local. A parent-owned control pipe ends it; no PID-pattern signals are used.
Wrapper preflight self-tests execute its actual entry point in check-only mode,
so invalid secret/endpoint text is never executed, even if a guard regresses.

Correction check record: targeted Go tests, all-fixture preflight, wrapper
negative/positive integration checks, and 46 contained flags/one-click CLI cases
completed with exit 0 in the strict call that records this section. Current
source was built before the CLI cases. Instrument-sensitivity proof remains
pending the checked plant runner; the validation stop remains until that proof
is rendered. Plant run 4 was incomplete: endpoint red exit was 50, not the
manifest's expected 1, and the following stub variant was unrun. The corrected
manifest records Scrut's measured exit 50. No incomplete run is called clean.

## Focused round 3 stop

All five aspects completed with exit 0 and validated results over snapshot
`57a2292ab3897bb93840f5638bd32b654e41ab64`; the authoritative snapshot value is
in `worker/evidence/reviews/focused3/scope.json`. Recapture found no lasting
changes. Code and types reported no findings. The loop is **stopped: recurring
defect**, pending the nullable-response class decision sent to TOOLING.

| Finding | Reviewer label | Our severity | Take or record | Review again |
| --- | --- | --- | --- | --- |
| Null JMAP invocations panic in query/get loops | important | Important: batch reporting can abort after completed POSTs | Take after nullable-response class approval; audit every dereferenced protocol member | Yes, malformed-response behavior; recurrence of unchecked null members |
| Signature tests omit only one coverage field | important | Important: destination-header signature guard could be weakened unnoticed | Take: independently omit each field and plant both conjuncts | No, tests only |
| Only one eligible confirmed sender tested | important | Important: the central batch continuation contract lacks coverage | Take: actual command with two distinct eligible senders, including first-POST failure | No, tests only |
| Credential fixture headings describe missing credentials | suggestion | Nit: headings mismatch the owned-fixture/refusing-default setup | Take: name failing credential setup accurately | No, docs only |

Counts remain structural 2 and focused 3 (Nit plus two Important rounds), trims 0,
Copilot 0. Approved behavioral fixes advance next to Critical-only round 4;
counts do not reset for this stop.

Plant run 5 ended incomplete: `command-consent` replaced the only use of
`confirm` with a literal, so red setup failed compilation. Later variants did
not run. The corrected mutant must retain a variable use (`confirm || !confirm`)
while forcing confirmation. No compilation failure is accepted as guard proof.
The CLI validation process stop therefore remains active.

## Nullable-response class correction

CONDUCTOR reply 000007 approved the complete nullable JMAP response-member
class audit. Validation now runs at the query/get response boundary before
selection loops consume protocol values. The response, invocation count,
invocation, operation-specific arguments and account must be valid. Query IDs
and get not-found IDs cannot be empty. Every returned message requires an ID
and received date; address and header objects cannot be null; header names must
be nonempty ASCII field names. Optional empty collections fail closed through
no-match or missing-header results. Typed method errors retain their cause.
Both production query and get paths call this boundary.

| Audited member | Boundary check / test |
| --- | --- |
| Response and invocation collection | Null response, empty/multiple results, null invocation |
| Query/get arguments | Null/typed-null arguments, decoded wire null, wrong operation/account |
| Method error | Null argument or missing error type |
| Query IDs / not-found IDs | Empty IDs rejected |
| Message / ID / date | Null message, empty ID, null received date rejected |
| From addresses | Null address rejected; empty collections cannot select a sender |
| Headers | Null header and invalid name rejected; missing required fields skip |
| Production boundary wiring | Independent query/get bypass mutants |

`TestSenderLookupInvalidResponses` covers the class through the production
lookup entry point. Batch tests exercise two eligible senders, continuation
after an HTTP failure, and retention of earlier POST results after later
metadata failure. Independent signature tests omit each required signed field.
Credential-fixture headings now describe the refusing setup accurately.

Correction checks: targeted client, unsubscribe and command Go packages exited
0 in the strict call recording this section. Plant manifest now declares 77
variants, including each new structural guard and both boundary wiring sites;
proof remains pending execution. Focused count stays 3; next is Critical-only
round 4. Structural loop remains clean at Important after 2 rounds. Trims 0,
Copilot 0. The nullable-response gate is resolved; the CLI validation process
stop still awaits rendered instrument proof.

## Focused round 4 and trims

All five focused aspects completed with validated output and exit 0, using the
same standalone read-only model/settings as earlier rounds. Snapshot
`1bd5b8b4ea95090f43d448a290191f6b64eae67e` recapture had no lasting changes and
unchanged scope. Code, errors, comments and types returned no findings.

| Finding | Reviewer label | Our severity | Take or record | Review again |
| --- | --- | --- | --- | --- |
| JSON tests decode using production DTO tags | suggestion | Nit: independent published-key coverage missing | Take: fixed key/type expectations in actual command tests and field-renaming plant | No, tests only |

End state: **clean at Critical**, focused count 4 (Nit, two Important, one
Critical). Structural count 2, clean at Important. No further behavior edits.

Both trim passes ran once, standalone read-only `gpt-6.1-sol`, low effort,
872000 window, with exit 0 and complete file-list/table reports. Comment trim
proposed no changes. Prose trim took mailto wording, deduplication wording and
removal of scope-check narration. Retained the proposed shortened introduction:
explicit and repeatable sender input explains the interface. No source comments
changed; unchanged-code comparison was unnecessary. Historical audit evidence,
commands, caveats and constraints stayed intact. No other proposals declined.
Raw trim reports are under `worker/evidence/reviews/trims/`.

The checked plant run 6 rendered successfully before subsequent test-only edits;
its generated record is `/private/tmp/fm88-plants-run6-ledger.md`. As authorized
by CONDUCTOR reply 000006, its actual-wrapper containment proofs lifted the CLI
validation process stop. The updated manifest adds the JSON contract plant;
final proof will be rerun over the updated test inputs.

`lint-and-fix --no-push`: initial lint exited 2 for unchecked test pipe closes
and unkeyed aliased header literals. Those test-only issues were corrected.
Final `make fmt`, `make lint`, affected command/unsubscribe tests, ShellCheck and
shfmt check all exited 0 in the strict call recording this section. No enforced
work skipped; no push. Advisory shell checks also completed. Final complete
`make test-ci`, final plants and exact-tree tidy remain pending.

## Handoff disposition

Final plant evidence is generated in
`docs/reviews/2026-11-18-one-click-plants.md`, rendered with exit 0 against the
updated committed inputs. It includes the independent JSON field-name plant.
Raw evidence and its generated status live under the worker mailbox
`worker/evidence/plants7/` and `worker/evidence/plants7-status.md`. Historical
review paths above are relative to that mailbox; its full location is
`/Users/ctm/.local/state/codex-mailboxes/tooling/88-one-click-unsubscribe/`.
The external review captures were archived there before their temporary copy
was removed. Earlier incomplete plant runs remain historical, not final proof.

Reviews ended structural **clean at Important** (2 rounds) and focused
**clean at Critical** (4 rounds); comments/prose trims each ran once. Copilot
rounds: 0. Optional signature-parser allocation and parser extraction suggestions
remain recorded, without behavior changes. Provider-delivery trust and the
imported/rewritten-message limitation are documented; no independent provenance
verification is claimed. Deferred work remains on the open issue.

The final repository-check receipt will be written to the mailbox path
`worker/evidence/final-gate/receipt.json`. That receipt and a same-call verify
against the final tree, rather than this paragraph, determine final check status.
The gate runs isolated `make test-ci`, shell checks, the containment self-test,
and current-input plant verification, then rebuilds three logical signed commits
and checks equality with the pre-tidy backup. No push, PR or merge is authorized
for this worker. The handoff must retain the early fake-token session-request
caveat above. No real unsubscribe request was sent.
