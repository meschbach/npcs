#!/bin/bash

compose=no
release=no
for arg in "$@" ; do
  case "$arg" in
    compose)
      compose=yes
      ARCHITECTURES="amd64 arm64"
      ;;
    release)
      release=yes
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
    for arch in $ARCHITECTURES; do
      echo -n "    $arch"
      CGO_ENABLED=0 GOOS=linux GOARCH=${arch} go build -trimpath -ldflags='-w -s -extldflags "-static"' -o "${output}_${arch}" ./$entry
      echo "."
    done
  fi
}

mkdir -p bin
compile cmd/competition bin/competition
compile cmd/simple bin/simple

echo "All programs compiled."

if [ "x$release" = "xyes" ]; then
  echo "Building docker images"
  docker buildx build --platform "linux/amd64,linux/arm64" -f ./cmd/competition/Dockerfile .
fi

if [ "$compose" = "yes" ]; then
  docker compose up --build
fi
