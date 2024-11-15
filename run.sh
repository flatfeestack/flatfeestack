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
Usage: $(basename "${BASH_SOURCE[0]}") [-h] [-na] [-ne] [-nb] [-nf] [-ns] [-sb] [-db] [-rm] [-ah] [-rh]

Build and run Flatfeestack.

Available options:
-h, --help          Print this help and exit
-ss, --skip-stripe  Don't setup stripe
-sb, --skip-build   Don't run docker-compose build
-rm, --remove-data  Remove the database and chain folder
EOF
  exit
}

parse_params() {
  # default values of variables set from params
  include_build=true
  include_stripe=true
  external=''
  internal="$PROJECTS"

  while :; do
    case "${1-}" in
    -h | --help) usage ;;
    --no-color) NO_COLOR=1 ;;
    -ss | --skip-stripe) external="${external} stripe-webhook"; internal="${internal//stripe-webhook/}"; include_stripe=false;;
    -sb | --skip-build) include_build=false;;
    -rm | --remove-data) sudo rm -rf .db .ganache .repos .chain .stripe;;
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
