#!/bin/bash

compose=no
release=no
for arg in "$@" ; do
  case "$arg" in
    compose)
      compose=yes
      OPERATING_SYSTEMS="linux darwin"
      ARCHITECTURES="amd64 arm64"
      ;;
    release)
      release=yes
      OPERATING_SYSTEMS="linux darwin"
      ARCHITECTURES="amd64 arm64"
      ;;
    *)
      echo "Unknown argument $arg"
      exit -1
      ;;
  esac
done

if [ "x$release" = "xyes" ]; then
  export ARCHITECTURES="amd64 arm64"
fi

set -e
function compile() {
  entry=$1; shift
  output=$1; shift

  echo "Compiling $output"
  if [ -z "$ARCHITECTURES" ]; then
    go build -o $output ./$entry
  else
    for os in $OPERATING_SYSTEMS; do
      for arch in $ARCHITECTURES; do
        echo -n "    $os-$arch"
        mkdir -p "bin/${os}-${arch}"
        CGO_ENABLED=0 GOOS=${os} GOARCH=${arch} go build -trimpath -ldflags='-w -s -extldflags "-static"' -o "bin/${os}-${arch}/${output}" ./$entry
        echo "."
      done
    done
  fi
}

mkdir -p bin
compile cmd/competition competition
compile cmd/simple simple

echo "All programs compiled."

if [ "$compose" = "yes" ]; then
  docker compose down
  docker compose up --build --attach-dependencies player_1 player_2
  docker compose down
fi
