#!/bin/bash

function compile() {
  entry=$1; shift
  output=$1; shift

  go build -o $output ./$entry
}

mkdir -p bin
compile cmd/competition bin/competition
