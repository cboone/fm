# One-click unsubscribe validation

## One-click rejects exact filters

```scrut
$ "$TESTDIR/../fm" unsubscribe --one-click --sender a@example.com --from-domain example.com 2>&1
{
  "error": "general_error",
  "message": "--one-click does not accept filter flags"
}
[1]
```

## One-click rejects a non-Fastmail session flag before authentication

```scrut
$ env -u FM_FORMAT -u FM_ACCOUNT_ID HOME=/nonexistent FM_CREDENTIAL_COMMAND='exit 99' $TESTDIR/../fm unsubscribe --one-click --sender a@example.com --session-url "$FM_TEST_SESSION_URL" 2>&1
{
  "error": "general_error",
  "message": "one-click requires the Fastmail session endpoint"
}
[1]
```

## One-click rejects a non-Fastmail session environment before authentication

```scrut
$ env -u FM_FORMAT -u FM_ACCOUNT_ID HOME=/nonexistent FM_CREDENTIAL_COMMAND='exit 99' FM_SESSION_URL="$FM_TEST_SESSION_URL" $TESTDIR/../fm unsubscribe --one-click --sender a@example.com 2>&1
{
  "error": "general_error",
  "message": "one-click requires the Fastmail session endpoint"
}
[1]
```

## One-click rejects conflicting consent flags before authentication

The credential command refuses authentication, so a validation regression cannot
contact a mailbox.

```scrut
$ env -u FM_FORMAT -u FM_ACCOUNT_ID HOME=/nonexistent FM_CREDENTIAL_COMMAND='exit 99' $TESTDIR/../fm unsubscribe --one-click --sender a@example.com --confirm --dry-run 2>&1
{
  "error": "general_error",
  "message": "use --one-click --sender addresses, without IDs, --draft or conflicting --confirm/--dry-run"
}
[1]
```

