# Credential command config review

The branch accepts `credential-command` as a config file alias for
`credential_command`. Both spellings must agree when present. Flags and
the environment retain priority over valid config values.

The base was `a23d957`. Each counted round compared the scope before
and after reviewing, with no scope or path changes. Read-only Codex
reviews used `gpt-6.1-sol`, high effort and a context window of 872000.

| Round                | Result                           | Evidence       |
| -------------------- | -------------------------------- | -------------- |
| Structural Nit       | Took one Important finding       | `structural-1` |
| Structural Important | Clean at Important               | `structural-2` |
| Focused Nit          | Ended: fixes changed no behavior | `focused-1`    |

- Structural Nit took the anti-pattern reviewer's Important finding.
  A YAML null could let the other key's command run despite a conflict.
  The fix checks whether each key is present and tests both null cases.
  The four cleanup aspects found no issues in either round.
- Focused review ran code, errors, comments and tests. No production
  type definitions changed. The tests reviewer suggested updating five
  stale plant fixtures. The finding counted as a Nit because it affected
  fixture replay. Ordinary tests still exercised the behavior.
  The audit compared every existing patch for the changed Go files
  against the base and branch. All five new application failures
  came from the config reader rename. The fixes update patch context.
- Synthetic CLI tests use temporary files, an isolated process environment
  and credential fallback stubs. The checked runner's final receipts
  and the scope captures are in the TOOLING handoff.
- Each trim profile ran once with `gpt-6.1-sol` at low effort and a
  context window of 872000. The comment pass removed a redundant
  function description. The prose pass took four edits.
