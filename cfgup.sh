#!/usr/bin/env bash

# based on https://betterdev.blog/minimal-safe-bash-script-template/
set -Eeuo pipefail
trap cleanup SIGINT SIGTERM ERR EXIT

cleanup() {
  trap - SIGINT SIGTERM ERR EXIT
  # script cleanup here
}

usage() {
  cat <<EOF
Usage: $(basename "${BASH_SOURCE[0]}") [-n name] [-m mapping] [-s service] [-c context] [-h] [--no-color]

Update config file for docker services.

Available options:

-h, --help      Print this help and exit
--no-color      Do not use colors in output
-n, --name      The name of the config as defined in "docker config create caddy_cfg Caddyfile". The name is here caddy_cfg
-m, --mapping   The mapping from source file to target file. E.g., env/backend.env.test:/app/.env
-s, --service   The name of the service. E.g., flatfee_backend
-c, --context   The docker context
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
  while :; do
    case "${1-}" in
    -h | --help) usage ;;
    --no-color) NO_COLOR=1 ;;
    -n | --name)
      NAME="${2-}"
      shift
      ;;
    -m | --mapping)
      MAPPING="${2-}"
      shift
      ;;
    -s | --service)
      SERVICE="${2-}"
      shift
      ;;
    -c | --context)
      CONTEXT="${2-}"
      shift
      ;;
    -?*) die "Unknown option: $1";;
    *) break ;;
    esac
    shift
  done

  args=("$@")
  return 0
}

parse_params "$@"
setup_colors

if [ -z "${NAME-}" ]; then
  die "${RED}Name not set"
fi
if [ -z "${MAPPING-}" ] ; then
  die "${RED}Mapping not set"
fi
if [ -z "${SERVICE-}" ]; then
  die "${RED}Service not set"
fi
if [ -z "${CONTEXT-}" ]; then
  die "${RED}Context not set"
fi

msg "${GREEN}Updating mapping in [$MAPPING] for [$NAME] with context [$CONTEXT] for service [$SERVICE]"

SRC=$(echo "$MAPPING"| cut -d : -f 1)
TARGET=$(echo "$MAPPING"| cut -d : -f 2)

docker --context "$CONTEXT" config create "${NAME}_tmp" "$SRC"
docker --context "$CONTEXT" service update --config-rm "$NAME" --config-add source="${NAME}_tmp",target="$TARGET" "$SERVICE"
docker --context "$CONTEXT" config rm "$NAME"
docker --context "$CONTEXT" config create "$NAME" "$SRC"
docker --context "$CONTEXT" service update --config-rm "${NAME}_tmp" --config-add source="$NAME",target="$TARGET" "$SERVICE"
docker --context "$CONTEXT" config rm "${NAME}_tmp"
docker --context "$CONTEXT" service update "$SERVICE" --force