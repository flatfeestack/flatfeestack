#!/usr/bin/env bash

# based on https://betterdev.blog/minimal-safe-bash-script-template/
# trap? -> https://vaneyckt.io/posts/safer_bash_scripts_with_set_euxo_pipefail/
set -Eeuo pipefail

script_dir=$(cd "$(dirname "${BASH_SOURCE[0]}")" &>/dev/null && pwd -P)

# check what machine we are on
case "$(uname -s)" in
    Linux*)     machine=Linux;;
    Darwin*)    machine=Mac;;
    CYGWIN*)    machine=Cygwin;;
    MINGW*)     machine=MinGw;;
    *)          machine="UNKNOWN"
esac

# depending on the machine extract the container host ip or use the term from docker
if [ $machine == Linux ]
then
	hostip=$(ifconfig docker0 | awk '/inet / {print $2}')
else
	hostip="host.docker.internal"
fi
export HOST_IP=$hostip

usage() {
  cat <<EOF
Usage: $(basename "${BASH_SOURCE[0]}") [-h] [-v] [-f] -p param_value arg1 [arg2...]

Build and run flatfeestack.

Available options:

-h, --help          Print this help and exit
-v, --verbose       Print script debug info
-na, --no-auth      Don't start auth
-ne, --no-engine    Don't start analysis-engine
-ni, --no-api       Don't start api
-ns, --no-scheduler Don't start scheduler
-np, --no-payout    Don't start payout
-db, --database     Run the DB instance only
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
  auth=''
  engine=''
  api=''
  scheduler=''
  payout=''
  db=''
  hosts='localhost:127.0.0.1' #we need to define a mapping, otherwise docker-compose complains

  while :; do
    case "${1-}" in
    -h | --help) usage ;;
    -v | --verbose) set -x ;;
    --no-color) NO_COLOR=1 ;;
    -na | --no-auth) auth='--scale auth=0' hosts="auth:${hostip}";;
    -ne | --no-engine) engine='--scale analysis-engine=0' hosts="analysis-engine:${hostip}";;
    -ni | --no-api) api='--scale api=0' hosts="api:${hostip}";;
    -ns | --no-scheduler) scheduler='--scale scheduler=0' hosts="scheduler:${hostip}";;
    -np | --no-payout) payout='--scale payout=0' hosts="payout:${hostip}";;
    -nf | --no-frontend) payout='--scale frontend=0' hosts="frontend:${hostip}";;
    -db | --database) db='db' ;;
    -?*) die "Unknown option: $1" ;;
    *) break ;;
    esac
    shift
  done

  args=("$@")

  return 0
}

parse_params "$@"
setup_colors

# script logic here
msg "${GREEN} build container"
docker-compose build --parallel
msg "${GREEN} run container"
HOSTS=${hosts} docker-compose up --abort-on-container-exit ${auth} ${engine} ${scheduler} ${payout} ${db}
