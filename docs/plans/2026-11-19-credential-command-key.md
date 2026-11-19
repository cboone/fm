# Credential command config spelling

[The issue][issue] reports that the CLI flag spelling, `credential-command`,
is silently ignored in a config file. On main at `a23d957`, a temporary YAML
file with that spelling fell through to the repository's credential stub.
The underscore spelling ran the synthetic command from the file.

Adopt the config map normalization from `547543db` for the credential
command. Read the file before binding higher priority sources so
validation sees both original values. Keep `credential_command` canonical
and accept `credential-command` as an alias. Different values for the two
spellings make the file invalid. The error names both keys without printing
their values. In a valid file, the CLI flag and environment still take
precedence in that order.

Exercise both spellings, matching and conflicting duplicates, and
overrides through the actual CLI with a temporary file and harmless
commands. Keep Keychain and account access outside validation.

[issue]: https://github.com/cboone/fm/issues/84
