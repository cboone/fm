# fm live integration tests

These tests require a valid credential command and explicit opt-in.
They are skipped automatically when preconditions are not met.

## Preconditions

Each test block below includes an inline precondition guard of the form:
`test "$FM_LIVE_TESTS" = "1" -a -n "$FM_CREDENTIAL_COMMAND" || exit 80`, so that
live tests are only run when `FM_LIVE_TESTS=1` and `FM_CREDENTIAL_COMMAND` is set.

## Session returns JSON with username

```scrut
$ test "$FM_LIVE_TESTS" = "1" -a -n "$FM_CREDENTIAL_COMMAND" || exit 80; $TESTDIR/../fm session --format json 2>&1
{
  "username": *, (glob)
* (glob+)
}
```

## Session with text format

```scrut
$ test "$FM_LIVE_TESTS" = "1" -a -n "$FM_CREDENTIAL_COMMAND" || exit 80; $TESTDIR/../fm session --format text 2>&1
Username: * (glob)
* (glob+)
```

## Mailboxes returns JSON array

```scrut
$ test "$FM_LIVE_TESTS" = "1" -a -n "$FM_CREDENTIAL_COMMAND" || exit 80; $TESTDIR/../fm mailboxes --format json 2>&1
[
* (glob+)
  "id": *, (glob)
* (glob+)
]
```

## Mailboxes roles-only entries all have role field

```scrut
$ test "$FM_LIVE_TESTS" = "1" -a -n "$FM_CREDENTIAL_COMMAND" || exit 80; $TESTDIR/../fm mailboxes --roles-only --format json | python3 -c 'import json,sys; data=json.load(sys.stdin); assert isinstance(data, list); assert data; assert all("role" in mb and mb["role"] for mb in data)'
```

## List returns JSON with total and emails

```scrut
$ test "$FM_LIVE_TESTS" = "1" -a -n "$FM_CREDENTIAL_COMMAND" || exit 80; $TESTDIR/../fm list --limit 1 --format json 2>&1
{
  "total": *, (glob)
  "offset": 0,
  "emails": [
* (glob+)
}
```

## List with text format includes Total and ID lines

```scrut
$ test "$FM_LIVE_TESTS" = "1" -a -n "$FM_CREDENTIAL_COMMAND" || exit 80; $TESTDIR/../fm list --limit 1 --format text 2>&1
Total: * (glob)
* ID: * (glob)
* (glob+)
```

## Search with filter returns results for known sender

```scrut
$ test "$FM_LIVE_TESTS" = "1" -a -n "$FM_CREDENTIAL_COMMAND" || exit 80; SENDER=$($TESTDIR/../fm list --limit 10 --format json | python3 -c 'import json,sys; d=json.load(sys.stdin); print(next((addr["email"] for e in d.get("emails", []) for addr in e.get("from", []) if addr.get("email")), ""))'); test -n "$SENDER" || exit 80; $TESTDIR/../fm search --from "$SENDER" --limit 1 --format json 2>&1
{
  "total": *, (glob)
  "offset": 0,
* (glob+)
}
```

## Search with text query returns snippets

```scrut
$ test "$FM_LIVE_TESTS" = "1" -a -n "$FM_CREDENTIAL_COMMAND" || exit 80; TERM=$($TESTDIR/../fm list --limit 10 --format json | python3 -c 'import json,sys,re; d=json.load(sys.stdin); print(next((w.lower() for e in d.get("emails", []) for w in re.findall(r"[A-Za-z]{3,}", e.get("subject", ""))), "the"))'); test -n "$TERM" || exit 80; $TESTDIR/../fm search "$TERM" --limit 1 --format json 2>&1
{
  "total": *, (glob)
  "offset": 0,
* (glob+)
      "snippet": *, (glob)
* (glob+)
}
```

## Read known email ID returns body

```scrut
$ test "$FM_LIVE_TESTS" = "1" -a -n "$FM_CREDENTIAL_COMMAND" || exit 80; EMAIL_ID=$($TESTDIR/../fm list --limit 1 --format json | python3 -c 'import json,sys; d=json.load(sys.stdin); emails=d.get("emails", []); print(emails[0]["id"] if emails else "")'); test -n "$EMAIL_ID" || exit 80; $TESTDIR/../fm read "$EMAIL_ID" --format json 2>&1
{
  "id": *, (glob)
* (glob+)
  "body": *, (glob)
* (glob+)
}
```

## Mailboxes text format shows tabular output

```scrut
$ test "$FM_LIVE_TESTS" = "1" -a -n "$FM_CREDENTIAL_COMMAND" || exit 80; $TESTDIR/../fm mailboxes --format text 2>&1
* total:* unread:* (glob)
* (glob+)
```

## Mark-read marks an email as read

```scrut
$ test "$FM_LIVE_TESTS" = "1" -a -n "$FM_CREDENTIAL_COMMAND" || exit 80; EMAIL_ID=$($TESTDIR/../fm list --unread --limit 1 --format json | python3 -c 'import json,sys; d=json.load(sys.stdin); emails=d.get("emails", []); print(emails[0]["id"] if emails else "")'); test -n "$EMAIL_ID" || exit 80; $TESTDIR/../fm mark-read "$EMAIL_ID" --format json 2>&1
{
  "matched": 1,
  "processed": 1,
  "failed": 0,
  "marked_as_read": [
    "*" (glob)
  ],
  "errors": []
}
```

## Mark-read removes the email from unread results

```scrut
$ test "$FM_LIVE_TESTS" = "1" -a -n "$FM_CREDENTIAL_COMMAND" || exit 80; EMAIL_ID=$($TESTDIR/../fm list --unread --limit 1 --format json | python3 -c 'import json,sys; d=json.load(sys.stdin); emails=d.get("emails", []); print(emails[0]["id"] if emails else "")'); test -n "$EMAIL_ID" || exit 80; $TESTDIR/../fm mark-read "$EMAIL_ID" --format json >/dev/null; $TESTDIR/../fm list --unread --limit 50 --format json | python3 -c 'import json,sys; target=sys.argv[1]; d=json.load(sys.stdin); ids={e.get("id") for e in d.get("emails", [])}; assert target not in ids, f"{target} is still unread"' "$EMAIL_ID"
```

## Error in text format uses Error prefix

```scrut
$ test "$FM_LIVE_TESTS" = "1" || exit 80; $TESTDIR/../fm session --format text --credential-command "echo bad-token" --session-url http://localhost:1/jmap 2>&1
Error [authentication_failed]: * (glob)
* (glob+)
[1]
```
