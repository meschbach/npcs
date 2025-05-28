#!/bin/bash

ver=${1:-dev}

set -e
outputs="darwin-arm64 linux-arm64 linux-amd64"
for output in ${outputs}; do
  target_file="npcs-${ver}-$output.tgz"
  echo $target_file
  tar zcf $target_file -C "bin/${output}" .
done
