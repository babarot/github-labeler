#!/bin/bash

args=()

if ${INPUT_IMPORT:-false}; then
  args+=("--import")
fi

if ${INPUT_DRYRUN:-false}; then
  args+=("--dry-run")
fi

config=${INPUT_CONFIG}
github-labeler --config=${config} ${args[@]}
