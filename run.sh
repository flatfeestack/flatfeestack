#!/usr/bin/env bash

# based on https://betterdev.blog/minimal-safe-bash-script-template/
set -Eeuo pipefail
trap cleanup SIGINT SIGTERM ERR EXIT

cleanup() {
  trap - SIGINT SIGTERM ERR EXIT
  # script cleanup here
  if [ -n "${PID-}" ]; then
    kill "$PID" 2>/dev/null
  fi
}

host_ip() {
  # check what machine we are on
  host_ip="localhost"
  case "$(uname -s)" in
      Linux*)     host_ip=$(ip -br -4 a show dev docker0|tr -s ' '|cut -d' ' -f 3|cut -d/ -f1);;
      Darwin*)    host_ip="host.docker.internal";;
      *)          host_ip="localhost";;
  esac
  export HOST_IP=$host_ip
  msg "${GREEN}Using [${host_ip}] as IP to reach docker containers";
}

usage() {
  cat <<EOF
Usage: $(basename "${BASH_SOURCE[0]}") [-h] [-na] [-ne] [-nb] [-np] [-npn] [-nf] [-sb] [-db] [-rm]

Build and run flatfeestack.

Available options:

-h, --help          Print this help and exit
-na, --no-auth      Don't start auth
-ne, --no-engine    Don't start analyzer engine
-nb, --no-backend   Don't start backend
-np, --no-payout    Don't start payout
-nf, --no-frontend  Dont' start frontend
-sb, --skip-build   Don't run docker-compose build (if your machine is slow)
-db, --db-only      Run the DB instance only, this ignores all the other options
-rm, --remove-data  Remove the database and chain folder
EOF
  exit
}

setup_colors() {
  if [[ -t 2 ]] && [[ -z "${NO_COLOR-}" ]] && [[ "${TERM-}" != "dumb" ]]; then
    NOFORMAT='\033[0m' RED='\033[0;31m' GREEN='\033[0;32m' ORANGE='\033[0;33m' BLUE='\033[0;34m' PURPLE='\033[0;35m' CYAN='\033[0;36m' YELLOW='\033[1;33m'
  else
    NOFORMAT='' RED='' GREEN='' ORANGE='' BLUE='' PURPLE='' CYAN='' YELLOW=''
  fi
}

msg() {
  echo >&2 -e "${1-}"
}

die() {
  local msg=$1
  local code=${2-1} # default exit status 1
  msg "$msg"
  exit "$code"
}

parse_params() {
  # default values of variables set from params
  include_build=true
  external=''
  internal='db caddy ganache auth analyzer backend payout frontend'

  while :; do
    case "${1-}" in
    -h | --help) usage ;;
    --no-color) NO_COLOR=1 ;;
    -na | --no-auth) external="${external} auth"; internal="${internal//auth/}";;
    -ne | --no-engine) external="${external} analyzer"; internal="${internal//analyzer/}";;
    -nb | --no-backend) external="${external} backend"; internal="${internal//backend/}";;
    -np | --no-payout) external="${external} payout"; internal="${internal//payout/}";;
    -nf | --no-frontend) external="${external} frontend"; internal="${internal//frontend/}";;
    -sb | --skip-build) include_build=false;;
    -db | --db-only) internal='db'; external='caddy ganache auth analyzer backend payout frontend'; break;; #if this is set everything else is ignored
    -rm | --remove-data) rm -rf .db .chain;;
    -?*) die "Unknown option: $1";;
    *) break ;;
    esac
    shift
  done

  args=("$@")
  return 0
}

run_cmd() {
  if [ -f .env ]; then
    source .env
    if [ -n "${CMD-}" ]; then
      msg "${GREEN}Running command [$CMD]"
      eval "$CMD" &
      PID=$!
    fi
  fi
}

parse_params "$@"
setup_colors

host_ip
run_cmd
mkdir -p .db .chain

# here we set hosts that can be used in docker-compose. For those hosts
# that are excluded, one wants to start it locally. Since we use docker
# DNS that resolves e.g, db to an IP, we need to resolve db to localhost
[ -z "${external}" ] && external="localhost:127.0.0.1" || external="${external}:${host_ip}"
msg "${GREEN}Setting DNS hosts to [${external}], started at $(date)"

if [ "$include_build" = true ]; then
  msg "${GREEN}Run: docker-compose build --parallel ${internal}"
  EXTRA_HOSTS="${external}" docker-compose build --parallel ${internal}
fi

# https://stackoverflow.com/questions/56844746/how-to-set-uid-and-gid-in-docker-compose
# https://hub.docker.com/_/postgres
msg "${GREEN}Run: docker-compose up --abort-on-container-exit ${internal}"
EXTRA_HOSTS="${external}" docker-compose up --abort-on-container-exit ${internal}
