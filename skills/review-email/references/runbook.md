# Email Review Runbook

Detailed operational procedures for email triage sessions.

## Session Start Checklist

1. Check today's date against the handoff document's last-updated date.
1. If different day: start from Phase 1 (new mail may have arrived overnight).
1. Verify connectivity with `fm session`.
1. Get the unread sender distribution with `fm summary --unread --format text`.
1. Review user preferences recorded in the handoff document, especially personal-message gate and active holds.
1. Start from the largest non-personal sender groups.

## Batch Operation Patterns

### Processing Sequence

For each sender cohort, follow this sequence:

1. `fm mark-read --from <sender>` to mark read
1. `fm unflag --from <sender>` to remove flags (if applicable)
1. `fm archive --from <sender>` to archive

For action commands (`archive`, `spam`, `mark-read`, `flag`, `unflag`, `move`), use `--dry-run` to preview changes before committing.

### Filter-Based Operations

Action commands support filter flags (`--from`, `--unread`, `--mailbox`, etc.) so you can operate on sender cohorts directly without collecting IDs first:

```bash
fm mark-read --from notifications@github.com
fm archive --from notifications@github.com
```

### Verification

After each batch, verify the result:

```bash
fm summary --unread --format text
```

Check that the processed sender no longer appears in the summary, or that the count dropped as expected.

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

## Draft Composition Patterns

Drafts are safe: they are created in the Drafts mailbox with the `$draft` keyword and cannot be sent by `fm`. The user must review and send drafts manually from Fastmail.

Always offer draft composition to the user rather than drafting autonomously. The user decides what to say and when to send.

### Reply

```bash
fm draft --reply-to <id> --body "Reply text"
```

### Reply All

```bash
fm draft --reply-all <id> --body "Reply text"
```

### Forward

```bash
fm draft --forward <id> --to <addr> --body "FYI - see below."
```

### Unsubscribe Request

When an unsubscribe link is unavailable or unreliable:

```bash
fm draft --reply-to <id> --body "Please remove me from this mailing list. Thank you."
```

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
