#!/bin/bash

set -e
compose_base=deploy/docker-compose

function cmd_platform_up() {
  (cd $compose_base/platform
  ./compose.sh up
  )
}

function cmd_dc1_up() {
  (cd $compose_base/dc1
  ./compose.sh up
  )
}

function cmd_up() {
    cmd_platform_up
    cmd_dc1_up
    (
    ./dev.sh compose
    )
}

function cmd_down() {
    docker-compose down --volumes
    (
      cd $compose_base/dc1
      ./compose.sh down
    )

    (
      cd $compose_base/platform
      ./compose.sh down
    )
}


function help() {
  echo "$0 <cmd>"
  echo "Where <cmd> is:"
  echo "    platform up - up"
}

function summary_help() {
  echo "$0 <subcommand>"
  echo "Where <subcommand> is:"
  echo "  * platform_up   - brings up the platform (integration postgres, OTLP system)"
}

cmd="$1" ; shift || true
function bad_subcommand() {
    echo "$cmd is an unknown subcommand"
    summary_help
}

case "$cmd" in
  platform_up)
    cmd_platform_up
    ;;
  dc1_up)
    cmd_dc1_up
    ;;
  down)
    cmd_down
    ;;
  *)
    cmd_up
    ;;
esac
