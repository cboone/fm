# Read and filter flag colors

Current `origin/main` at `bc9962f` writes flag colors but cannot read or filter
them. See [the flag color issue][issue].

1. Use the existing flag color bits to decode JMAP keywords and build exact
   server-side filters for `list` and `search`.
1. Show the decoded color in JSON and text. A flagged message with all three
   bits clear reads as red. Unflagged messages and the reserved bit pattern
   have a null color.
1. Test the command paths with synthetic JMAP requests and document the flags
   and the red convention in help and the README.
1. Run the full local review order with before-and-after scope proof for every
   round. Validate the final signed tree with a checked-gate receipt and
   open a ready pull request.

The worktree uses Go 1.24.7 from `go.mod` and checksum-verified macOS builds
of the CI versions of Scrut and golangci-lint.

[issue]: https://github.com/cboone/fm/issues/90
