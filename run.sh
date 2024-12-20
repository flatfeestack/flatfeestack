#!/usr/bin/env bash

PROJECTS='db caddy ganache auth analyzer backend frontend-svelte5 stripe-webhook'

# Improved version based on https://betterdev.blog/minimal-safe-bash-script-template/
set -Eeuo pipefail
trap 'cleanup $?' SIGINT SIGTERM ERR EXIT

cleanup() {
  trap - SIGINT SIGTERM ERR EXIT
  # sAdd any cleanup tasks here
}

setup_colors() {
  if [[ -t 2 ]] && [[ -z "${NO_COLOR-}" ]] && [[ "${TERM-}" != "dumb" ]]; then
    NOFORMAT='\033[0m' RED='\033[0;31m' GREEN='\033[0;32m' ORANGE='\033[0;33m'
  else
    NOFORMAT='' RED='' GREEN='' ORANGE=''
  fi
}

msg() {
  echo >&2 -e "${1-}"
}

die() {
  msg "${RED}${1-}${NOFORMAT}"
  exit "${2-1}"
}

usage() {
  cat <<EOF
Usage: $(basename "${BASH_SOURCE[0]}") [-h] [-ss] [-sb] [-sd] [-rm]

Build and run FlatFeeStack.

Available options:
-h, --help               Print this help and exit
-ss, --skip-stripe       Don't setup stripe
-sb, --skip-build        Don't run docker-compose build
-sd, --skip-start-docker Don't try to starte docker
-rm, --remove-data       Remove the database and chain folder
EOF
  exit
}

parse_params() {
  # default values of variables set from params
  include_build=true
  include_stripe=true
  include_docker_start=true
  external=''
  internal="$PROJECTS"

  while :; do
    case "${1-}" in
    -h | --help) usage ;;
    --no-color) NO_COLOR=1 ;;
    -ss | --skip-stripe) external="${external} stripe-webhook"; internal="${internal//stripe-webhook/}"; include_stripe=false;;
    -sb | --skip-build) include_build=false;;
    -sd | --skip-start-docker) include_docker_start=false;;
    -rm | --remove-data) sudo rm -rf .db .ganache .repos .chain .stripe;;
    -rmdb | --remove-db) sudo rm -rf .db;;
    -?*) die "Unknown option: $1";;
    *) break ;;
    esac
    shift
  done

  #args=("$@")
  return 0
}

setup_colors
parse_params "$@"

# on linux, check if command exists, and start docker if not running
# also make sure that no other containers are running, since we start all relevant containers
# here. Use -sd to skip this check, or if you have containers, that you want to keep running
if [ "$include_docker_start" = true ] && command -v systemctl >/dev/null 2>&1; then
  if ! systemctl is-active --quiet docker; then
    msg "Docker is not running. Starting Docker service..."
    sudo systemctl start docker

    # Wait for Docker to fully start
    while ! systemctl is-active --quiet docker; do
        sleep 1
    done
    msg "${GREEN}Docker service started successfully${NOFORMAT}"
  fi

  # Get all running containers and stop them
  running_containers=$(docker ps -q)
  if [ -n "$running_containers" ]; then
    msg "Stopping all running containers..."
    docker stop $running_containers
    msg "${GREEN}All containers stopped${NOFORMAT}"
  else
    msg "${GREEN}No running containers found${NOFORMAT}"
  fi
fi

if [ "$include_stripe" = true ]; then
  msg "${GREEN}Setup Stripe${NOFORMAT}"
  docker compose build stripe-setup
  docker compose up stripe-setup
else
  msg "${ORANGE}Skip Stripe setup${NOFORMAT}"
fi

if [ "$include_build" = true ]; then
  msg "${GREEN}Run: docker compose build ${internal}${NOFORMAT}"
  docker compose build ${internal}
else
  msg "${ORANGE}Skip: docker compose build ${internal}${NOFORMAT}"
fi

# https://stackoverflow.com/questions/56844746/how-to-set-uid-and-gid-in-docker-compose
# https://hub.docker.com/_/postgres
msg "${GREEN}Run: docker compose up --abort-on-container-exit ${internal}${NOFORMAT}"
docker compose up --abort-on-container-exit ${internal}
