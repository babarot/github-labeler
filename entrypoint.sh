#!/bin/bash

manifest=${INPUT_MANIFEST}

import=${INPUT_IMPORT:-false}
if ${import}; then
  github-labeler --import --config=${manifest}
  exit ${?}
fi

github-labeler --config=${manifest}
