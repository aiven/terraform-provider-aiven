#!/usr/bin/env bash

# Run tests and determine if failures should be retried
# Usage: ./run-with-smart-retry.sh

set -o pipefail

# Exit codes
SUCCESS_CODE=0
RETRY_ERROR_CODE=42     # This is the exit code to tell the retry action to look for
NO_RETRY_ERROR_CODE=1

# Error patterns that should trigger retries
declare -a RETRY_PATTERNS=(
  "text file busy"
)

# Pipe output to tee so that it can be both displayed and captured
OUTPUT=$(make test-acc 2>&1 | tee >(cat >&2))
EXIT_CODE=${PIPESTATUS[0]}

# If tests succeeded, exit with success
if [ $EXIT_CODE -eq 0 ]; then
  exit $SUCCESS_CODE
fi

# Tests failed, check if it matches any retry patterns
for pattern in "${RETRY_PATTERNS[@]}"; do
  if echo "$OUTPUT" | grep -q "$pattern"; then
    echo "Found retry pattern: '$pattern' - Will trigger retry"
    exit $RETRY_ERROR_CODE  # Exit with special code that signals retry
  fi
done

exit $NO_RETRY_ERROR_CODE  # Don't retry