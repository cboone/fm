# jm help output

Verify that the root help output shows all commands and key information.

## Root help

```scrut
$ $TESTDIR/../jm --help
jm is a command-line tool for reading, searching, and triaging email (glob)
via the JMAP protocol. It connects to Fastmail (or any JMAP server) and (glob)
provides read, search, archive, and spam operations. (glob)
 (regex)
Sending and deleting email are structurally disallowed. (glob)
 (regex)
Usage: (glob)
  jm [command] (glob)
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
Use "jm [command] --help" for more information about a command. (glob)
```

## Session command help

```scrut
$ $TESTDIR/../jm session --help
Display JMAP session info (verify connectivity and auth) (glob)
 (regex)
Usage: (glob)
  jm session [flags] (glob)
* (glob+)
```

## Mailboxes command help

```scrut
$ $TESTDIR/../jm mailboxes --help
List all mailboxes (folders/labels) in the account (glob)
 (regex)
Usage: (glob)
  jm mailboxes [flags] (glob)
 (regex)
Flags: (glob)
*--help* (glob)
*--roles-only* (glob)
* (glob*)
```

## List command help

```scrut
$ $TESTDIR/../jm list --help
List emails in a mailbox (glob)
 (regex)
Usage: (glob)
  jm list [flags] (glob)
 (regex)
Flags: (glob)
*--flagged* (glob)
*--help* (glob)
*--limit* (glob)
*--mailbox* (glob)
*--offset* (glob)
*--sort* (glob)
*--unflagged* (glob)
*--unread* (glob)
* (glob*)
```

## Read command help

```scrut
$ $TESTDIR/../jm read --help
Read the full content of an email (glob)
 (regex)
Usage: (glob)
  jm read <email-id> [flags] (glob)
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
$ $TESTDIR/../jm search --help
Search emails using full-text search and/or structured filters. (glob)
The optional [query] argument searches across subject, from, to, and body. (glob)
If omitted, only the provided flags/filters are used for matching. (glob)
 (regex)
Usage: (glob)
  jm search [query] [flags] (glob)
 (regex)
Flags: (glob)
*--after* (glob)
*--before* (glob)
*--flagged* (glob)
*--from* (glob)
*--has-attachment* (glob)
*--help* (glob)
*--limit* (glob)
*--mailbox* (glob)
*--offset* (glob)
*--sort* (glob)
*--subject* (glob)
*--to* (glob)
*--unflagged* (glob)
*--unread* (glob)
* (glob*)
```

## Archive command help

```scrut
$ $TESTDIR/../jm archive --help
Move emails to the Archive mailbox (glob)
 (regex)
Usage: (glob)
  jm archive <email-id> [email-id...] [flags] (glob)
* (glob+)
```

## Spam command help

```scrut
$ $TESTDIR/../jm spam --help
Move emails to the Junk/Spam mailbox (glob)
 (regex)
Usage: (glob)
  jm spam <email-id> [email-id...] [flags] (glob)
* (glob+)
```

## Mark-read command help

```scrut
$ $TESTDIR/../jm mark-read --help
Mark emails as read (set the $seen keyword) (glob)
 (regex)
Usage: (glob)
  jm mark-read <email-id> [email-id...] [flags] (glob)
* (glob+)
```

## Flag command help

```scrut
$ $TESTDIR/../jm flag --help
Flag emails (set the $flagged keyword) (glob)
 (regex)
Usage: (glob)
  jm flag <email-id> [email-id...] [flags] (glob)
* (glob+)
```

## Unflag command help

```scrut
$ $TESTDIR/../jm unflag --help
Unflag emails (remove the $flagged keyword) (glob)
 (regex)
Usage: (glob)
  jm unflag <email-id> [email-id...] [flags] (glob)
* (glob+)
```

## Move command help

```scrut
$ $TESTDIR/../jm move --help
Move one or more emails to a target mailbox (by name or ID). (glob)
Moving to Trash or Deleted Items is not permitted. (glob)
 (regex)
Usage: (glob)
  jm move <email-id> [email-id...] --to <mailbox> [flags] (glob)
 (regex)
Flags: (glob)
*--help* (glob)
*--to* (glob)
* (glob*)
```
