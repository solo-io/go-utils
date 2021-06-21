#!/bin/bash

set -ex

set +e

make go-fmt
if [[ $? -ne 0 ]]; then
  echo "Code formatting failed"
  exit 1;
fi
if [[ $(git status --porcelain | wc -l) -ne 0 ]]; then
  echo "Error: Generating code produced a non-empty diff"
  echo "Try running 'make install-go-tools go-fmt -B' then re-pushing."
  git status --porcelain
  git diff | cat
  exit 1;
fi