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
Usage: $(basename "${BASH_SOURCE[0]}") [-p] [-m] [-h] [--no-color]

Update config file for docker services.

Available options:

-h, --help      Print this help and exit
--no-color      Do not use colors in output
-p, --patch     Increase the patch version -> 1.0.4 to 1.0.5
-m, --minor     Increase the minor version -> 1.0.4 to 1.1.0
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
  ACTION=""
  while :; do
    case "${1-}" in
    -h | --help) usage ;;
    --no-color) NO_COLOR=1 ;;
    -p | --patch)ACTION="patch";;
    -m | --minor)ACTION="minor";;
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

command -v semver >/dev/null || {
  die "${RED}Command 'semver' not found in \$PATH. Please, first install it${RED}"
}

if [ -z "${ACTION-}" ]; then
  die "${RED}You need to set either minor or patch"
fi

msg "${GREEN}Tagging and deploying..."
git fetch --tags
CURRENT=$(git tag --sort=v:refname | tail -1)
NEXT=$(semver "$CURRENT" -i $ACTION)
msg "${GREEN}Deploy from $CURRENT to $NEXT"
git tag "$NEXT"
git push origin "$NEXT"
