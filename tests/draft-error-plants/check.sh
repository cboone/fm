#!/usr/bin/env bash
# Run synthetic responses without credential helpers or network transports.
set -euo pipefail

function main() {
  ./draft-error-tests -test.v \
    -test.run 'TestCreateDraft_(SetErrorDetails|WithoutSubmission)$|TestGetAllIdentities_SubmissionCapability$'
}

main "$@"
