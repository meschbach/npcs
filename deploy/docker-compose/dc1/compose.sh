#!/bin/bash

export COMPOSE_PROJECT_NAME=npcs_dc1

set -e
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
  echo -n "Waiting for PostgreSQL to be ready..."
  local retries=0
  local max_attempts=10
  while ! PGPASSWORD="primary_dc1_password" "$libpq_dir/bin/psql" -h localhost -p 16004 -U primary_dc1_user -d primary_dc1_db -c '\q' 2>/dev/null; do
    echo -n "."
    sleep 1
    ((retries++))
    if [ $retries -ge $max_attempts ]; then
      echo "Failed to connect to PostgreSQL after $max_attempts attempts"
      exit 1
    fi
  done
  echo "PostgreSQL is ready!"
}

function cmd_run() {
  if docker network ls | grep -q "npcs_dc1_internal"; then
    version=$( docker network inspect npcs_dc1_internal| jq -r '.[0].Labels["com.meschbach/version"]' )
    if [ $version -lt 4 ] ; then
      container_name="${COMPOSE_PROJECT_NAME}-pg_primary-1"
      docker stop "$container_name"
      docker rm  "$container_name" --volumes || echo "WARNING: Unable to remove old docker contained.  Continuing"
      docker volume rm "${COMPOSE_PROJECT_NAME}_pg_primary_data"
    fi
  fi
  
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
    echo "$cmd is a valid subcommand"
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
    bad_subcommand
    ;;
esac
