# fm sieve commands

Verify sieve subcommand help output and error handling.

## Sieve parent help

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

## Sieve list help

```scrut
$ $TESTDIR/../fm sieve list --help
List all sieve scripts (glob)
 (regex)
Usage: (glob)
  fm sieve list [flags] (glob)
* (glob+)
```

## Sieve show help

```scrut
$ $TESTDIR/../fm sieve show --help
Show a sieve script's metadata and content (glob)
 (regex)
Usage: (glob)
  fm sieve show <script-id> [flags] (glob)
* (glob+)
```

## Sieve create help

```scrut
$ $TESTDIR/../fm sieve create --help
Create a new sieve script on the server. (glob)
* (glob+)
Usage: (glob)
  fm sieve create --name <name> [flags] (glob)
 (regex)
Flags: (glob)
*--action* (glob)
*--activate* (glob)
*-n, --dry-run* (glob)
*--fileinto* (glob)
*--from * (glob)
*--from-domain* (glob)
*--help* (glob)
*--name* (glob)
*--script-stdin* (glob)
* (glob*)
```

## Sieve activate help

```scrut
$ $TESTDIR/../fm sieve activate --help
Activate a sieve script by ID. (glob)
* (glob+)
Usage: (glob)
  fm sieve activate <script-id> [flags] (glob)
 (regex)
Flags: (glob)
*-n, --dry-run* (glob)
*--help* (glob)
* (glob*)
```

## Sieve deactivate help

```scrut
$ $TESTDIR/../fm sieve deactivate --help
Deactivate the currently active sieve script (glob)
 (regex)
Usage: (glob)
  fm sieve deactivate [flags] (glob)
 (regex)
Flags: (glob)
*-n, --dry-run* (glob)
*--help* (glob)
* (glob*)
```

## Sieve delete help

```scrut
$ $TESTDIR/../fm sieve delete --help
Delete a sieve script by ID. (glob)
* (glob+)
Usage: (glob)
  fm sieve delete <script-id> [flags] (glob)
 (regex)
Flags: (glob)
*-n, --dry-run* (glob)
*--help* (glob)
* (glob*)
```

## Sieve validate help

```scrut
$ $TESTDIR/../fm sieve validate --help
Validate sieve script syntax on the server without creating a script. (glob)
* (glob+)
Usage: (glob)
  fm sieve validate [flags] (glob)
 (regex)
Flags: (glob)
*--help* (glob)
*--script * (glob)
*--script-stdin* (glob)
* (glob*)
```

## Sieve list without credential command

```scrut
$ env -u FM_CREDENTIAL_COMMAND -u FM_SESSION_URL -u FM_FORMAT -u FM_ACCOUNT_ID HOME=/nonexistent $TESTDIR/../fm sieve list 2>&1
{
  "error": "authentication_failed",
  "message": "credential command failed: *", (glob)
  "hint": "Check your credential command or the token it returns"
}
[1]
```

## Sieve show without credential command

```scrut
$ env -u FM_CREDENTIAL_COMMAND -u FM_SESSION_URL -u FM_FORMAT -u FM_ACCOUNT_ID HOME=/nonexistent $TESTDIR/../fm sieve show S1 2>&1
{
  "error": "authentication_failed",
  "message": "credential command failed: *", (glob)
  "hint": "Check your credential command or the token it returns"
}
[1]
```

## Sieve show requires an argument

```scrut
$ $TESTDIR/../fm sieve show 2>&1
Error: accepts 1 arg(s), received 0
[1]
```

## Sieve create requires name

```scrut
$ env -u FM_CREDENTIAL_COMMAND -u FM_SESSION_URL -u FM_FORMAT -u FM_ACCOUNT_ID HOME=/nonexistent $TESTDIR/../fm sieve create --from "a@b.com" --action junk 2>&1
{
  "error": "general_error",
  "message": "required flag \"name\" not set",
  "hint": "Provide a name with --name"
}
[1]
```

## Sieve create requires template flags or stdin

```scrut
$ env -u FM_CREDENTIAL_COMMAND -u FM_SESSION_URL -u FM_FORMAT -u FM_ACCOUNT_ID HOME=/nonexistent $TESTDIR/../fm sieve create --name "test" 2>&1
{
  "error": "general_error",
  "message": "provide either template flags (--from/--from-domain + --action) or --script-stdin"
}
[1]
```

## Sieve activate requires an argument

```scrut
$ $TESTDIR/../fm sieve activate 2>&1
Error: accepts 1 arg(s), received 0
[1]
```

## Sieve delete requires an argument

```scrut
$ $TESTDIR/../fm sieve delete 2>&1
Error: accepts 1 arg(s), received 0
[1]
```

## Sieve validate requires script input

```scrut
$ env -u FM_CREDENTIAL_COMMAND -u FM_SESSION_URL -u FM_FORMAT -u FM_ACCOUNT_ID HOME=/nonexistent $TESTDIR/../fm sieve validate 2>&1
{
  "error": "general_error",
  "message": "either --script or --script-stdin is required"
}
[1]
```

## Sieve create dry-run shows generated script

```scrut
$ env -u FM_CREDENTIAL_COMMAND -u FM_SESSION_URL -u FM_FORMAT -u FM_ACCOUNT_ID HOME=/nonexistent $TESTDIR/../fm sieve create --name "Block spam" --from "spam@example.com" --action junk --dry-run 2>&1
{
  "operation": "create",
  "script": "Block spam",
  "content": "require [\"fileinto\"];\n\nif address :is \"from\" \"spam@example.com\" {\n    fileinto \"Junk\";\n    stop;\n}\n"
}
```
