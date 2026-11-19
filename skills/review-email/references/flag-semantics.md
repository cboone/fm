# Flag Semantics

Color flag meanings and usage patterns for email triage.

## Color Meanings

| Color  | Meaning                              | Examples                                |
| ------ | ------------------------------------ | --------------------------------------- |
| Red    | Sensitive admin or personal items    | IRS notices, legal documents, identity  |
| Orange | Upcoming commitments or events       | Event tickets, reservations, deadlines  |
| Yellow | Receipts or reference documents      | Payment confirmations, shipping notices |
| Green  | Stored value                         | Gift cards, credits, certificates       |
| Gray   | Low-priority or someday items        | Things to review later, no urgency      |
| Blue   | (Available for project-specific use) |                                         |
| Purple | (Available for project-specific use) |                                         |

## Flag Commands

Set a color flag:

```bash
fm flag --color <color> <email-id>
```

Show a color in `list`, `search` or `read`. Filter by color with
`--flagged --color <color>` in `list` or `search`.

Clear the color bits while keeping the flag:

```bash
fm unflag --color <email-id>
```

A flagged message with all bits clear reads as red and matches the red filter.

Fully unflag (remove flag and color):

```bash
fm unflag <email-id>
```

## Stale Flag Cleanup

Check for orphaned flags in non-inbox mailboxes:

```bash
fm summary --flagged -m Archive --format text
fm summary --flagged -m Junk --format text
```

Users may archive or spam flagged emails from their email client without removing the flag first. Unflag any found in Archive or Junk.

## When to Flag

- Flag for user review when you cannot confidently categorize an email.
- Flag actionable items using the appropriate color.
- Do not flag emails that will be immediately archived or spammed.
