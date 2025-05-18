#!/bin/bash

set -e
function compile() {
  entry=$1; shift
  output=$1; shift

  echo "Compiling $output"
  go build -o $output ./$entry
}

mkdir -p bin
compile cmd/competition bin/competition
compile cmd/simple bin/simple

echo "All programs compiled."
