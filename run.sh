#!/usr/bin/env bash

# Improved version based on https://betterdev.blog/minimal-safe-bash-script-template/
set -Eeuo pipefail
trap 'cleanup $?' SIGINT SIGTERM ERR EXIT

PROJECTS='db caddy ganache auth analyzer backend frontend stripe-webhook'

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

manage_hosts() {
    local hostnamesToAdd=("$@")
    IFS=' ' read -r -a hostnamesToRemove <<< "$PROJECTS"

    for hostname in "${hostnamesToRemove[@]}"; do
      if [[ "$(uname)" == "Darwin" ]]; then
        sudo sed -i '' "/[[:space:]]$hostname$/d" /etc/hosts
      else
        sudo sed -i "/[[:space:]]$hostname$/d" /etc/hosts
      fi
    done

    for hostname in "${hostnamesToAdd[@]}"; do
      if ! grep -q "[[:space:]]$hostname$" /etc/hosts; then
        echo "127.0.0.1 $hostname" | sudo tee -a /etc/hosts > /dev/null
      fi
    done
}

check_hosts() {
  local hostnames=("$@")
  for hostname in "${hostnames[@]}"; do
    if ! grep -q "$hostname" /etc/hosts; then
      msg "${ORANGE}Not found: $hostname in /etc/hosts. Please run: $(basename "${BASH_SOURCE[0]}") --add-hosts${NOFORMAT}"
    fi
  done
}

usage() {
  cat <<EOF
Usage: $(basename "${BASH_SOURCE[0]}") [-h] [-na] [-ne] [-nb] [-nf] [-ns] [-sb] [-db] [-rm] [-ah] [-rh]

Build and run Flatfeestack.

Available options:
-h, --help          Print this help and exit
-na, --no-auth      Don't start auth
-ne, --no-engine    Don't start analyzer engine
-nb, --no-backend   Don't start backend
-nf, --no-frontend  Don't start frontend
-ns, --no-stripe    Don't start stripe-webhook
-sb, --skip-build   Don't run docker-compose build
-db, --db-only      Run the DB instance only
-rm, --remove-data  Remove the database and chain folder
-ah, --add-hosts    Add project hostnames to /etc/hosts
-rh, --remove-hosts Remove project hostnames from /etc/hosts
EOF
  exit
}

parse_params() {
  # default values of variables set from params
  include_build=true
  external=''
  internal="$PROJECTS"
  add_hosts=false;
  remove_hosts=false;

  while :; do
    case "${1-}" in
    -h | --help) usage ;;
    --no-color) NO_COLOR=1 ;;
    -na | --no-auth) external="${external} auth"; internal="${internal//auth/}";;
    -ne | --no-engine) external="${external} analyzer"; internal="${internal//analyzer/}";;
    -nb | --no-backend) external="${external} backend"; internal="${internal//backend/}";;
    -nf | --no-frontend) external="${external} frontend"; internal="${internal//frontend/}";;
    -ns | --no-stripe) external="${external} stripe-webhook"; internal="${internal//stripe-webhook/}";;
    -sb | --skip-build) include_build=false;;
    -db | --db-only) internal='db'; external="${PROJECTS//db}";; #if this is set everything else is ignored
    -ah | --add-hosts) add_hosts=true;;
    -rh | --remove-hosts) remove_hosts=true;;
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

if [ "$remove_hosts" = true ]; then
  manage_hosts
elif [ "$add_hosts" = true ]; then
  manage_hosts $external
fi

check_hosts $external

#if there is no login yet, it will prompt for one
docker compose build stripe-setup
docker compose up stripe-setup

if [ "$include_build" = true ]; then
  msg "${GREEN}Run: docker compose build ${internal}${NOFORMAT}"
  docker compose build ${internal}
fi

# here we set hosts that can be used in docker-compose. For those hosts
# that are excluded, one wants to start it locally. Since we use docker
# DNS that resolves e.g, db to an IP, we need to resolve db to localhost
external="${external} localhost:127.0.0.1"
msg "${GREEN}Setting DNS hosts to [${external}], started at $(date)${NOFORMAT}"

# https://stackoverflow.com/questions/56844746/how-to-set-uid-and-gid-in-docker-compose
# https://hub.docker.com/_/postgres
msg "${GREEN}Run: docker compose up --abort-on-container-exit ${internal}${NOFORMAT}"
EXTRA_HOSTS="${external}" docker compose up --abort-on-container-exit ${internal}
