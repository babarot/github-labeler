#!/bin/bash

main() {
  local -a args=()

  if ${INPUT_IMPORT:-false}; then
    args+=("--import")
  fi

  if ${INPUT_DRYRUN:-false}; then
    args+=("--dry-run")
  fi

  local config=${INPUT_CONFIG}
  github-labeler --config=${config} ${args[@]}
}

set -o pipefail

main "$@" | tee -a result
result="$(cat result)"

# https://github.community/t5/GitHub-Actions/set-output-Truncates-Multiline-Strings/td-p/37870
echo "::set-output name=result::${result//$'\n'/'%0A'}"
