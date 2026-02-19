# fm help output

Verify that the root help output shows all commands and key information.

## Root help

```scrut
$ $TESTDIR/../fm --help
fm is a command-line tool for reading, searching, triaging, and drafting Fastmail (glob)
email via the JMAP protocol. It connects to Fastmail (or any JMAP server) and (glob)
provides read, search, archive, spam, and draft operations. (glob)
 (regex)
Sending and deleting email are structurally disallowed. (glob)
 (regex)
Usage: (glob)
  fm [command] (glob)
 (regex)
Available Commands: (glob)
  archive * (glob)
  completion * (glob)
  draft * (glob)
  flag * (glob)
  help * (glob)
  list * (glob)
  mailboxes * (glob)
  mark-read * (glob)
  move * (glob)
  read * (glob)
  search * (glob)
  session * (glob)
  sieve * (glob)
  spam * (glob)
  stats * (glob)
  summary * (glob)
  unflag * (glob)
 (regex)
Flags: (glob)
* (glob+)
 (regex)
Use "fm [command] --help" for more information about a command. (glob)
```

## Session command help

```scrut
$ $TESTDIR/../fm session --help
Display JMAP session info (verify connectivity and auth) (glob)
 (regex)
Usage: (glob)
  fm session [flags] (glob)
* (glob+)
```

## Mailboxes command help

```scrut
$ $TESTDIR/../fm mailboxes --help
List all mailboxes (folders/labels) in the account (glob)
 (regex)
Usage: (glob)
  fm mailboxes [flags] (glob)
 (regex)
Flags: (glob)
*--help* (glob)
*--roles-only* (glob)
* (glob*)
```

## List command help

```scrut
$ $TESTDIR/../fm list --help
List emails in a mailbox (glob)
 (regex)
Usage: (glob)
  fm list [flags] (glob)
 (regex)
Flags: (glob)
*-f, --flagged* (glob)
*--help* (glob)
*-l, --limit* (glob)
*-m, --mailbox* (glob)
*-o, --offset* (glob)
*-s, --sort* (glob)
*--unflagged* (glob)
*-u, --unread* (glob)
* (glob*)
```

## Read command help

```scrut
$ $TESTDIR/../fm read --help
Read the full content of an email (glob)
 (regex)
Usage: (glob)
  fm read <email-id> [flags] (glob)
 (regex)
Flags: (glob)
*--help* (glob)
*--html* (glob)
*--raw-headers* (glob)
*--thread* (glob)
* (glob*)
```

## Search command help

```scrut
$ $TESTDIR/../fm search --help
Search emails using full-text search and/or structured filters. (glob)
The optional [query] argument searches across subject, from, to, and body. (glob)
If omitted, only the provided flags/filters are used for matching. (glob)
 (regex)
Usage: (glob)
  fm search [query] [flags] (glob)
 (regex)
Flags: (glob)
*--after* (glob)
*--before* (glob)
*-f, --flagged* (glob)
*--from* (glob)
*--has-attachment* (glob)
*--help* (glob)
*-l, --limit* (glob)
*-m, --mailbox* (glob)
*-o, --offset* (glob)
*-s, --sort* (glob)
*--subject* (glob)
*--to* (glob)
*--unflagged* (glob)
*-u, --unread* (glob)
* (glob*)
```

## Stats command help

```scrut
$ $TESTDIR/../fm stats --help
Aggregate emails by sender address and display per-sender counts. (glob)
Queries all matching emails in the mailbox and groups them by sender, (glob)
sorted by volume descending. (glob)
 (regex)
Usage: (glob)
  fm stats [flags] (glob)
 (regex)
Flags: (glob)
*-f, --flagged* (glob)
*--help* (glob)
*-m, --mailbox* (glob)
*--subjects* (glob)
*--unflagged* (glob)
*-u, --unread* (glob)
* (glob*)
```

## Summary command help

```scrut
$ $TESTDIR/../fm summary --help
Aggregate emails by sender and domain, count unread messages, and optionally (glob)
detect newsletters. Provides a single-pass triage overview of a mailbox. (glob)
 (regex)
Usage: (glob)
  fm summary [flags] (glob)
 (regex)
Flags: (glob)
*-f, --flagged* (glob)
*--help* (glob)
*-l, --limit* (glob)
*-m, --mailbox* (glob)
*--newsletters* (glob)
*--subjects* (glob)
*--unflagged* (glob)
*-u, --unread* (glob)
* (glob*)
```

## Archive command help

```scrut
$ $TESTDIR/../fm archive --help
Move emails to the Archive mailbox (glob)
 (regex)
Usage: (glob)
  fm archive [email-id...] [flags] (glob)
 (regex)
Flags: (glob)
*--after* (glob)
*--before* (glob)
*-n, --dry-run* (glob)
*-f, --flagged* (glob)
*--from* (glob)
*--has-attachment* (glob)
*--help* (glob)
*-m, --mailbox* (glob)
*--subject* (glob)
*--to* (glob)
*--unflagged* (glob)
*-u, --unread* (glob)
* (glob*)
```

## Spam command help

```scrut
$ $TESTDIR/../fm spam --help
Move emails to the Junk/Spam mailbox (glob)
 (regex)
Usage: (glob)
  fm spam [email-id...] [flags] (glob)
 (regex)
Flags: (glob)
*--after* (glob)
*--before* (glob)
*-n, --dry-run* (glob)
*-f, --flagged* (glob)
*--from* (glob)
*--has-attachment* (glob)
*--help* (glob)
*-m, --mailbox* (glob)
*--subject* (glob)
*--to* (glob)
*--unflagged* (glob)
*-u, --unread* (glob)
* (glob*)
```

## Mark-read command help

```scrut
$ $TESTDIR/../fm mark-read --help
Mark emails as read (set the $seen keyword) (glob)
 (regex)
Usage: (glob)
  fm mark-read [email-id...] [flags] (glob)
 (regex)
Flags: (glob)
*--after* (glob)
*--before* (glob)
*-n, --dry-run* (glob)
*-f, --flagged* (glob)
*--from* (glob)
*--has-attachment* (glob)
*--help* (glob)
*-m, --mailbox* (glob)
*--subject* (glob)
*--to* (glob)
*--unflagged* (glob)
*-u, --unread* (glob)
* (glob*)
```

## Flag command help

```scrut
$ $TESTDIR/../fm flag --help
Flag emails (set the $flagged keyword) (glob)
 (regex)
Usage: (glob)
  fm flag [email-id...] [flags] (glob)
 (regex)
Flags: (glob)
*--after* (glob)
*--before* (glob)
*-c, --color* (glob)
*-n, --dry-run* (glob)
*-f, --flagged* (glob)
*--from* (glob)
*--has-attachment* (glob)
*--help* (glob)
*-m, --mailbox* (glob)
*--subject* (glob)
*--to* (glob)
*--unflagged* (glob)
*-u, --unread* (glob)
* (glob*)
```

## Unflag command help

```scrut
$ $TESTDIR/../fm unflag --help
Unflag emails (remove the $flagged keyword) (glob)
 (regex)
Usage: (glob)
  fm unflag [email-id...] [flags] (glob)
 (regex)
Flags: (glob)
*--after* (glob)
*--before* (glob)
*-c, --color* (glob)
*-n, --dry-run* (glob)
*-f, --flagged* (glob)
*--from* (glob)
*--has-attachment* (glob)
*--help* (glob)
*-m, --mailbox* (glob)
*--subject* (glob)
*--to* (glob)
*--unflagged* (glob)
*-u, --unread* (glob)
* (glob*)
```

## Move command help

```scrut
$ $TESTDIR/../fm move --help
Move one or more emails to a target mailbox (by name or ID). (glob)
Moving to Trash or Deleted Items is not permitted. (glob)
 (regex)
Usage: (glob)
  fm move [email-id...] --to <mailbox> [flags] (glob)
 (regex)
Flags: (glob)
*--after* (glob)
*--before* (glob)
*-n, --dry-run* (glob)
*-f, --flagged* (glob)
*--from* (glob)
*--has-attachment* (glob)
*--help* (glob)
*-m, --mailbox* (glob)
*--subject* (glob)
*--to* (glob)
*--unflagged* (glob)
*-u, --unread* (glob)
* (glob*)
```

## Sieve command help

```scrut
$ $TESTDIR/../fm sieve --help
Manage sieve filtering scripts on the server. (glob)
* (glob+)
Usage: (glob)
  fm sieve [command] (glob)
 (regex)
Available Commands: (glob)
  activate * (glob)
  create * (glob)
  deactivate * (glob)
  delete * (glob)
  list * (glob)
  show * (glob)
  validate * (glob)
* (glob+)
```

## Draft command help

```scrut
$ $TESTDIR/../fm draft --help
Create a draft email for later review and sending from Fastmail. (glob)
* (glob+)
Usage: (glob)
  fm draft [flags] (glob)
 (regex)
Flags: (glob)
*--bcc* (glob)
*--body * (glob)
*--body-stdin* (glob)
*--cc* (glob)
*--forward* (glob)
*--help* (glob)
*--html* (glob)
*--reply-all* (glob)
*--reply-to* (glob)
*--subject* (glob)
*--to* (glob)
* (glob*)
```
