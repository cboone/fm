# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

## [0.3.0] - 2026-03-27

### Added

- Sieve script management: list, show, create, validate, activate, deactivate, and delete
- Draft composition with reply, reply-all, forward, and new message modes
- Unsubscribe command with List-Unsubscribe header parsing (#70)
- `--from` flag for draft command with JMAP identity resolution (#69)
- `--subject` filter for list and summary commands (#72)
- To and CC fields in search and list text output (#73)
- Compact text output for action commands (#71)
- Filter flags (`--from`, `--to`, `--subject`, etc.) for bulk action commands
- Color flag support in flag/unflag commands
- Sender aggregation stats command
- Summary command for inbox triage
- List-Unsubscribe header in read output
- Configurable credential command, replacing static API token
- Terminal cell-width calculation for column alignment
- Review-email skill for guided inbox triage

### Changed

- Migrate goreleaser config from brews to homebrew_casks (#62)
- Migrate CI workflows to reusable gh-actions workflows
- Add golangci-lint to CI pipeline
- Refactor ListEmails to use ListOptions struct
- Add gitleaks secret scanning workflow

### Fixed

- Parse List-Unsubscribe headers case-insensitively
- Reject multiple email IDs in single-email commands
- Handle nil checks in actionVerb for correct zero-success output
- Decouple destination check from verb string
- Preserve newest sender display name in stats
- Default ListEmails sort and limit when unset
- Reject positional arguments for summary command
- Populate Name field in SieveDeleteResult
- Stop global flags section parsing at next heading or rule
- Remove undeclared GITHUB_TOKEN from release workflow
- Configure scrut build command for CLI tests

[unreleased]: https://github.com/cboone/fm/compare/v0.3.0...HEAD
[0.3.0]: https://github.com/cboone/fm/compare/v0.2.0...v0.3.0
