#!/bin/bash

set -x

make format-code
if [[ $? -ne 0 ]]; then
  echo "Code formatting failed"
  exit 1;
fi
if [[ $(git status --porcelain | wc -l) -ne 0 ]]; then
  echo "Error: Generating code produced a non-empty diff"
  echo "Try running 'make install-go-tools format-code -B' then re-pushing."
  git status --porcelain
  git diff | cat
  exit 1;
fi