#!/bin/bash

manifest=${INPUT_MANIFEST}

diff=${INPUT_DIFF:-false}
if ${diff}; then
  github-labeler --diff --config=${manifest}
  case ${?} in
    0)
      result=no
      ;;
    *)
      result=yes
      ;;
  esac
  echo ::set-output name=diff::"${result}"
  exit 0
fi

import=${INPUT_IMPORT:-false}
if ${import}; then
  github-labeler --import --config=${manifest}
  exit ${?}
fi

github-labeler --config=${manifest}
