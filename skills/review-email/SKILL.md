---
name: review-email
description: >-
  Guided email triage workflow using the fm CLI. Walks through an 8-phase
  inbox review process with safety gates for personal messages, batch
  operations by sender cohort, and session handoff tracking. Use when the
  user says "review email", "review my email", "triage email", "check my
  inbox", "email triage", "process my inbox", "triage my inbox", "clean
  up my inbox", "go through my email", "inbox zero", "email review", or
  any variant involving reviewing, triaging, or cleaning up Fastmail
  email.
---

# Review Email

Guided email triage using the `fm` CLI. Processes the inbox in priority order: personal messages first, then spam, promotional, transactional, newsletters, community mail, and stale flags.

## Prerequisites

This skill requires a **handoff document** in the consuming project. This is a Markdown file (typically `docs/fastmail/handoff.md`) that tracks session state between triage runs. It contains:

- `Last updated` date
- Current unread and flagged totals
- Flagged inbox message list
- Active rules and holds (sender/domain)
- Remaining unread landscape (categorized sender distribution)
- Recent session log

If the handoff document does not exist, create it on the first session start. See `./references/runbook.md` for the session end protocol that keeps it current.

## Core Principles

1. **Personal-message gate**: Surface direct personal communication before bulk actions. If uncertain, ask the user.
1. **Respect active holds**: Honor sender and domain holds recorded in the project's handoff document.
1. **Batch safely**: Work in sender cohorts, verify post-action counts after each batch.
1. **Keep history current**: Update the handoff document after meaningful work.
1. **Report fm issues**: When you encounter obstacles, missing features, bugs, or unexpected behavior in `fm`, tell the user and file a GitHub issue on `cboone/fm`.

## Workflow

### 1. Session Start

Check today's date against the `Last updated` date in the project's handoff document. If it is a different day, treat this as a new session and start from Phase 1.

Verify connectivity:

```bash
fm session
```

### 2. Get the Landscape

```bash
fm stats --unread --format text
```

Review the sender distribution. Identify high-volume, low-risk groups and any senders that might be personal.

### 3. Personal-Message Gate

Before processing any batch, spot-check subjects and previews for personal signals:

```bash
fm list --unread --limit 50 --format text
```

For ambiguous messages, inspect the full content:

```bash
fm read <id>
```

Apply the personal vs. bulk heuristics from `./references/runbook.md` to classify ambiguous messages. If uncertain, surface to the user before processing.

### 4. Process by Phase

Work through the triage phases in order. See `./references/triage-phases.md` for detailed phase descriptions.

**Phase sequence:**

1. High-priority messages (personal, time-sensitive)
1. Spam removal
1. Unwanted promotional email
1. Transactional cleanup
1. Newsletters worth reading
1. Community and organization email
1. Flagged email review for staleness
1. Ongoing maintenance

### 5. Batch Operations

For each sender cohort:

1. Preview with `--dry-run`:

```bash
fm archive --from <sender> --dry-run
```

2. Execute the batch:

```bash
fm mark-read --from <sender>
fm archive --from <sender>
```

3. Verify remaining counts:

```bash
fm stats --unread --format text
```

"Archive" implies mark-read and unflag first, then archive. "Spam" implies mark-read and unflag first, then spam.

See `./references/runbook.md` for detailed batch operation patterns.

### 6. Clean Up Stale Flags

Check for and remove orphaned flags in non-inbox mailboxes. See `./references/flag-semantics.md` for the cleanup procedure and the distinction between fully unflagging (`fm unflag <id>`) and removing only the color (`fm unflag --color <id>`).

### 7. Session End

Update the project's handoff document in place:

- Update the unread total and last-updated date
- Update the flagged inbox list if it changed
- Update active rules and holds if they changed
- Update the remaining unread landscape
- Append a session summary to the recent session log

Git history preserves previous snapshots automatically.

## Reference Navigation

**Triage phases:**

- `./references/triage-phases.md` - Detailed description of each phase with specific actions

**Operational procedures:**

- `./references/runbook.md` - Batch safety, personal vs. bulk heuristics, flag workflows, session checklists

**Flag semantics:**

- `./references/flag-semantics.md` - Color flag meanings and usage patterns
