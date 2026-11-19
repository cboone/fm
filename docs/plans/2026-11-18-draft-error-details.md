# Draft error details

Preserve the server's SetError type and description when draft creation fails.
Use the existing draft and identity client boundaries; no new tool is needed.
Draft creation must continue to work without submission. Identity lookup names
the missing submission capability and Email submission token scope, including
its sending permission, only when the authenticated session lacks it.

1. Extend synthetic draft tests for type-only, type plus description,
   description-only, empty and null rejections, and mail-only draft success.
2. Add synthetic identity tests for missing and present submission capability.
   Preserve cached identity behavior and existing transport error wrapping.
3. Plant regressions at each changed behavior using the installed checked runner.
4. Run the full local review order, then formatting, lint, unit tests and tidy.
   Reviews use separate read-only Codex runs, model `gpt-6.1-sol`, high effort,
   context window 872000. Skip CLI checks needing the one-click containment
   harness, as TOOLING authorized in mailbox reply 000001.
5. Fetch main before the final gate; integrate the one-click work if merged.
   Push and open a ready PR, noting unrun CLI checks, then report to TOOLING.
   Never merge.

No credentials, real mail, real drafts, accounts, live endpoints, installs,
live configuration, releases or tags. Scope excludes sender selection and
automatic identity fallback from issue 89. The initial check record documents
the current defect and why missing submission cannot explain every rejection.
