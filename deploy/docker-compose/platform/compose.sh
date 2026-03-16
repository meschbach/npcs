#!/bin/bash

libpq_dir=/opt/homebrew/opt/libpq
psql="$libpq_dir/bin/psql"

function ensure_nc() {
  # Check if required tools are installed
  if ! command -v nc &> /dev/null; then
     echo "Error: netcat is not installed. Please install it first."
     exit 1
  fi
}

function ensure_psql() {
 if ! command -v "$psql" &> /dev/null; then
     echo "Error: "$psql" is not installed. Please install first."
     exit 1
 fi
}

function help() {
  echo "$0 <cmd>"
  echo "Where <cmd> is:"
  echo "    up - updates the platform"
  echo "    remove - shutsdown and removes all containers related to this infrastructure"
}

function cmd_unknow_help() {
  echo "command '$cmd' is not found"
  help
}

function wait_for_postgres() {
  echo "Waiting for PostgreSQL to be ready..."
  while ! PGPASSWORD="integ-tests-password" "$libpq_dir/bin/psql" -h localhost -p 16003 -U integ_tests -d integ_db -c '\q' 2>/dev/null; do
    echo -n "."
    sleep 1
  done
  echo "PostgreSQL is ready!"
}

function cmd_run() {
  docker-compose up -d --remove-orphans
  ensure_psql
  wait_for_postgres
}

function cmd_pause() {
  docker-compose stop
}

function cmd_remove() {
    docker-compose down --volumes
}

function summary_help() {
  echo "$0 <subcommand>"
  echo "Where <subcommand> is:"
  echo "  * up   - attempts to place the containers into a running state (alias: 'run')"
  echo "  * pause - stops all containers from this group"
  echo "  * remove - removes the containers and networks.  volumes are retained"
  echo "  * down - removes all traces of this project"
}

cmd="$1" ; shift || true
function bad_subcommand() {
    echo "$cmd is an unknown subcommand"
    summary_help
}

case "$cmd" in
  run)
    cmd_run
    ;;
  up)
    cmd_run
    ;;
  pause)
    cmd_pause
    ;;
  remove)
    cmd_remove
    ;;
  down)
    cmd_remove
    ;;
  *)
    cmd_unknown_help
    ;;
esac
