#!/bin/bash

config=${INPUT_CONFIG}

import=${INPUT_IMPORT:-false}
if ${import}; then
  github-labeler --import --config=${config}
  exit ${?}
fi

github-labeler --config=${config}
