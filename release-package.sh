#!/bin/bash

set -e
outputs="darwin-arm64 linux-arm64 linux-amd64"
for output in ${outputs}; do
  tar zcvf $output.tgz -C "bin/${output}" .
done
