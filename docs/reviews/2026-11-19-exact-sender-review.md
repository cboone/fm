# Exact sender filter review

The branch adds case-insensitive address and ASCII domain filters to search
and action queries. JMAP From text search narrows candidates. The client
checks parsed addresses before counting, paging or returning IDs. The plan
records the approved selection strategy.

## Review rounds

The structural round and first focused round used base `bc9962f`. The second
focused round used `3702adfe`, which includes the merged flag color change.
Scope captures, per-aspect replies and comparisons are in the TOOLING
handoff. The read-only aspects used `gpt-6.1-sol`, high effort and a context
window of 872000. Each comparison required no scope or path changes.

| Round             | End state                        | Evidence       |
| ----------------- | -------------------------------- | -------------- |
| Structural Nit    | Ended: fixes changed no behavior | `structural-1` |
| Focused Nit       | Took one Important and one Nit   | `focused-1`    |
| Focused Important | Clean at Important               | `focused-2`    |

- Structural took the anti-pattern Nit about the first-match doc comment.
  Took four cleanup suggestions as Nits. Kept filter construction inside
  the selector, shared snippet attachment, reordered messages in place and
  removed redundant ASCII validation. The supported selection and output
  behavior stayed the same.
- Focused Nit took the tests reviewer's Important finding as Important.
  Losing either CLI option could widen an action query through another
  supplied filter. Synthetic CLI tests now exercise both flags on search and
  archive preview. Took the suggestion on summary batching as a Nit and
  added multi-batch, ordered results with distinct snippets.
- Main merged with no production conflicts. Resolved overlapping CLI
  documentation and Scrut snapshot additions. The synthetic CLI test
  verifies the exact domain and green flag conditions in the same query.
- Focused Important took a test suggestion as a Nit. New tests pin
  rejection of an explicitly empty address before authentication in search
  and archive. The test addition did not start another round.

## Validation

The final receipt and generated plant status are in the TOOLING handoff.
The plants cover candidate narrowing, comparison, pagination, ordered
selection, query-state checks, batching, snippets and CLI propagation. The
fixtures exercise local synthetic JMAP only. The checked runner restores
each planted input before rechecking it.

The final gate checks the tidied tree against the pre-tidy tree, runs the
repository CI command and verifies the plants. Scoped Markdown and shell
checks cover the new process records and plant scripts. Product Markdown
lint and formatting remain advisory where the baseline reports issues.

## Trims

Comment and prose passes used gpt-6.1-sol at low effort with a context window
of 872000. The comment pass proposed no edits. Both prose proposals were
taken with wording that preserves the candidate-check and pagination contract.
Historical line references in the initial check point to the revision recorded
there. The review evidence is under `trim` in the TOOLING handoff.
