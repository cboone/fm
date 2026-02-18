# Triage Phases

Work through these phases in order. Earlier phases surface urgent items; later phases reduce ongoing volume.

## Phase 1: High-Priority Messages

Scan the inbox for messages that need attention or action.

- Identify personal and direct messages and surface them to the user.
- Identify time-sensitive items: renewals, overdue notices, appointment reminders, expiring verification links.
- Flag high-priority items using the appropriate color (red for sensitive/admin, orange for upcoming commitments).

```bash
fm list --unread --limit 50 --format text
fm read <id>
fm flag --color red <id>
fm flag --color orange <id>
```

## Phase 2: Spam

Identify and remove actual spam.

- Look for senders that are clearly unsolicited or fraudulent.
- Mark as spam (which implies mark-read and unflag first).

```bash
fm mark-read --from <sender>
fm unflag --from <sender>
fm spam --from <sender>
```

## Phase 3: Unwanted Promotional Email

Identify promotional emails and other low-value mail that is not technically spam but is not worth reading.

- For each sender, assess how to stop future emails:
  - Unsubscribe link (preferred; `fm read <id>` exposes List-Unsubscribe when present).
  - Fastmail-side filter or rule if unsubscribe is unavailable or unreliable.
- Flag unsubscribe links for the user to handle from their client, then archive the messages.

```bash
fm read <id>
fm mark-read --from <sender>
fm archive --from <sender>
```

## Phase 4: Transactional Cleanup

Process receipts, payment confirmations, security alerts, and other automated account emails.

- Verify contents are expected and legitimate.
- Flag if actionable (green for stored value, yellow for receipts worth keeping).
- Archive the rest.

```bash
fm flag --color green <id>
fm flag --color yellow <id>
fm mark-read --from <sender>
fm archive --from <sender>
```

## Phase 5: Newsletters Worth Reading

Separate newsletters the user actually wants from the noise.

- Leave unread in inbox for the user to read, or flag and move per user preference.
- Ask the user about unfamiliar newsletter senders before archiving.

## Phase 6: Community and Organization Email

Handle emails from groups, communities, and organizations the user participates in.

- Not spam, not personal, but not high priority either.
- Surface anything with action items; archive the rest.

```bash
fm list --unread --from <group-address> --format text
fm mark-read --from <group-address>
fm archive --from <group-address>
```

## Phase 7: Flagged Email Review

Review existing flagged messages for staleness.

- Run the stale flag cleanup procedure from `../flag-semantics.md`.
- Check whether remaining flagged items are still relevant or have been resolved.
- Resolve, re-flag, or unflag as appropriate.

## Phase 8: Ongoing Maintenance

Reduce future inbox volume with structural changes.

- For repeat senders that keep appearing after unsubscribe, recommend Fastmail-side rules.
- Periodically revisit the triage plan and flag semantics as needs change.
