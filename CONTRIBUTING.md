## Running Tests

This repository contains a script for integration tests (`tests/assert.sh`).  
These tests verify that the library compiles down to a no-op with no references to `dlg` when the build tag is omitted.  

Because of these extra checks and the use of build tags, running `go test` on its own will not execute every test case.  
Instead use the provided `Makefile`:  
`make test` runs all Go unit tests + the integration tests.

## Pre-commit Hook

If you're developing in an environment where bash is available, you may use this script as a git pre-commit hook (.git/hooks/pre-commit):

```bash
#!/usr/bin/env bash

set -euo pipefail

# Format code
if command -v gofmt &>/dev/null; then
  gofmt -w ./
fi

typeset -i errors=0

# Checking if build tags are formatted properly
while IFS=  read -r f; do
  first_line="$(sed -n '1p' "${f}")"
  if [[ "$first_line" =~ ^//.*dlg.*$ ]]; then
    if [[ ! "$first_line" =~ ^//go:build\ [^\ ] ]]; then
      printf 'Error: Malformed build tag.\n'
      printf '  File:     %s\n' "${f}"
      printf '  Found:    %s\n' "${first_line}"
      printf '  Expected: //go:build <expression>\n'
      errors=$(( errors + 1))
    fi
  fi
done < <(git diff --cached --name-only --diff-filter=ACMR)

if (( errors )); then
    echo 'Commit rejected'
    exit 1
fi

exit 0
```
