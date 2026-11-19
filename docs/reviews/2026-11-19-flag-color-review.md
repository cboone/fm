# Flag color review

Base `bc9962f` supports writes only. No prior fix surfaced in the local and
remote branches or GitHub pull requests. The feature lets users audit
color use, including red flags whose three color keywords are absent.
The approved behavior follows RFC 9979. Read-only output reports null for
unflagged messages and the reserved pattern. The shared email summary
also carries the new field in dry-run JSON.

## Review results

All rounds used the fixed merge base. Each before-and-after comparison
exited 0 and returned `scope_changed: false`, `changed: []` and
`outside_scope: []`.

| Review                 | Result                           |
| ---------------------- | -------------------------------- |
| Structural Nit         | Ended: fixes changed no behavior |
| Focused Nit            | Two Important findings addressed |
| Focused Important      | Clean at Important               |
| Comment and prose trim | Comment pass and two prose edits |

- Structural Nit took the schema documentation finding and adopted two
  cleanup suggestions. List and search share query construction, and
  keyword spellings and patch paths are prepared once.
- Focused Nit took both findings as Important. Synthetic CLI tests verify
  color reaches both query entry points. All three text color writes
  propagate errors.
- Focused Important took two test suggestions as Nits. List and search
  assert explicit JSON nulls, and thread tests distinguish the selected
  message from other messages. Test and plant changes alone required
  no new review round.
- The trim pass kept the comment that explains lowercase JMAP keywords,
  shortened the color instructions, and moved the branch search result
  from the plan to this record.

## Validation

The plant manifest covers exact query bits, red and null decoding, CLI
argument propagation, independent thread colors and formatter write errors.
The checked-gate receipt and generated plant summary are in the TOOLING
handoff. Fixture data uses synthetic JMAP responses and a synthetic token.

The first final gate found a CLI test dependency on an earlier test's Viper
format override. CLI tests assert JSON nulls through the synthetic server.
The summary formatter test checks text output independently.

Product Markdown uses the surrounding formatting. Standalone Markdown
tools report advisory formatting issues on the baseline product docs.
