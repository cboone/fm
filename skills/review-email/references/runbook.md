# Email Review Runbook

Detailed operational procedures for email triage sessions.

## Session Start Checklist

1. Check today's date against the handoff document's last-updated date.
1. If different day: start from Phase 1 (new mail may have arrived overnight).
1. Verify connectivity with `fm session`.
1. Get the unread sender distribution with `fm stats --unread --format text`.
1. Review user preferences recorded in the handoff document, especially personal-message gate and active holds.
1. Start from the largest non-personal sender groups.

## Batch Operation Patterns

### Processing Sequence

For each sender cohort, follow this sequence:

1. `mark-read --from <sender>` to mark read
1. `unflag --from <sender>` to remove flags (if applicable)
1. `archive --from <sender>` to archive

Use `--dry-run` on any command to preview before committing.

### Filter-Based Operations

Action commands support filter flags (`--from`, `--unread`, `--mailbox`, etc.) so you can operate on sender cohorts directly without collecting IDs first:

```bash
fm mark-read --from notifications@github.com
fm archive --from notifications@github.com
```

### Verification

After each batch, verify the result:

```bash
fm stats --unread --format text
```

Check that the processed sender no longer appears in the unread stats, or that the count dropped as expected.

## Batch Safety

- Process by sender cohorts, not broad keyword sweeps.
- Use the verification step after each batch (remaining count by sender).
- Keep risky or sensitive senders for explicit user review.
- Use extra caution with financial, legal, medical, or identity-related content.

## Personal vs. Bulk Heuristics

### Likely Personal

- Conversational subject (`Re:` etc.) and direct `to` user address
- Sender is an individual and not a known list or newsletter sender
- Content is specific to a private thread (non-broadcast tone)

### Likely Bulk

- `to`/`cc` contains Google Groups or list addresses
- Recurring marketing cadence and promotional language
- Digest, status, or update style with templated structure

If uncertain, ask the user before processing.

## Archive and Spam Semantics

- "Archive" implies: mark-read, unflag, then archive.
- "Spam" implies: mark-read, unflag, then spam.

These are multi-step sequences, not single commands. Run each step explicitly.

## Flag Workflow

See `../flag-semantics.md` for color meanings, flag commands, and the stale flag cleanup procedure.

## Session End Protocol

Update the handoff document in place with:

- The updated unread total and last-updated date
- The updated flagged inbox list (if it changed)
- Any changes to active rules or holds
- The updated remaining unread landscape
- A session summary appended to the recent session log

The handoff document is a single living document. Git history preserves previous snapshots automatically. Do not create dated copies.
