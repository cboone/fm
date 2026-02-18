# fm help output

Verify that the root help output shows all commands and key information.

## Root help

```scrut
$ $TESTDIR/../fm --help
fm is a command-line tool for reading, searching, and triaging Fastmail email (glob)
via the JMAP protocol. It connects to Fastmail (or any JMAP server) and (glob)
provides read, search, archive, and spam operations. (glob)
 (regex)
Sending and deleting email are structurally disallowed. (glob)
 (regex)
Usage: (glob)
  fm [command] (glob)
 (regex)
Available Commands: (glob)
  archive * (glob)
  completion * (glob)
  flag * (glob)
  help * (glob)
  list * (glob)
  mailboxes * (glob)
  mark-read * (glob)
  move * (glob)
  read * (glob)
  search * (glob)
  session * (glob)
  spam * (glob)
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

## Archive command help

```scrut
$ $TESTDIR/../fm archive --help
Move emails to the Archive mailbox (glob)
 (regex)
Usage: (glob)
  fm archive <email-id> [email-id...] [flags] (glob)
 (regex)
Flags: (glob)
*-n, --dry-run* (glob)
*--help* (glob)
* (glob*)
```

## Spam command help

```scrut
$ $TESTDIR/../fm spam --help
Move emails to the Junk/Spam mailbox (glob)
 (regex)
Usage: (glob)
  fm spam <email-id> [email-id...] [flags] (glob)
 (regex)
Flags: (glob)
*-n, --dry-run* (glob)
*--help* (glob)
* (glob*)
```

## Mark-read command help

```scrut
$ $TESTDIR/../fm mark-read --help
Mark emails as read (set the $seen keyword) (glob)
 (regex)
Usage: (glob)
  fm mark-read <email-id> [email-id...] [flags] (glob)
 (regex)
Flags: (glob)
*-n, --dry-run* (glob)
*--help* (glob)
* (glob*)
```

## Flag command help

```scrut
$ $TESTDIR/../fm flag --help
Flag emails (set the $flagged keyword) (glob)
 (regex)
Usage: (glob)
  fm flag <email-id> [email-id...] [flags] (glob)
 (regex)
Flags: (glob)
*-c, --color* (glob)
*-n, --dry-run* (glob)
*--help* (glob)
* (glob*)
```

## Unflag command help

```scrut
$ $TESTDIR/../fm unflag --help
Unflag emails (remove the $flagged keyword) (glob)
 (regex)
Usage: (glob)
  fm unflag <email-id> [email-id...] [flags] (glob)
 (regex)
Flags: (glob)
*-c, --color* (glob)
*-n, --dry-run* (glob)
*--help* (glob)
* (glob*)
```

## Move command help

```scrut
$ $TESTDIR/../fm move --help
Move one or more emails to a target mailbox (by name or ID). (glob)
Moving to Trash or Deleted Items is not permitted. (glob)
 (regex)
Usage: (glob)
  fm move <email-id> [email-id...] --to <mailbox> [flags] (glob)
 (regex)
Flags: (glob)
*-n, --dry-run* (glob)
*--help* (glob)
*--to* (glob)
* (glob*)
```
