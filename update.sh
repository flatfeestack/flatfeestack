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
Usage: $(basename "${BASH_SOURCE[0]}") [-h] [--no-color] [-p] [-d] [-t]
Build and run Flatfeestack.
Available options:
-h, --help            Print this help and exit
--no-color            Disable color in console
-p, --pat             Use a personal access token (PAT) to clone the repo
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

parse_params() {
  pat=''

  while :; do
    case "${1-}" in
    -h | --help) usage ;;
    --no-color) NO_COLOR=1 ;;
    -p | --pat) pat="${2-}"; shift ;;
    -?*) die "Unknown option: $1" ;;
    *) break ;;
    esac
    shift
  done

  args=("$@")
  return 0
}

check_update() {
  git fetch
  BRANCH=$(git rev-parse --abbrev-ref HEAD)
  HEADHASH=$(git rev-parse HEAD)
  UPSTREAMHASH=$(git rev-parse $BRANCH@{upstream})
  if [ "$HEADHASH" != "$UPSTREAMHASH" ]; then
    msg "${BLUE}Updating the current repo first${NOFORMAT}"
    git pull
    msg "${GREEN}Running $0 again${NOFORMAT}"
    exec "$@"
    exit 0
  else
    msg "${GREEN}Repo is up to date${NOFORMAT}"
  fi
}

git_clone(){
  START_URL="git@github.com:"
  START_URL_CONSOLE="${START_URL}"
  if [ -n "${pat-}" ]; then
    START_URL="https://${pat}@github.com/"
    START_URL_CONSOLE="https://[PAT]@github.com/" #do not display sensitive information in the console
  fi

  msg "${GREEN}Cloning ${START_URL_CONSOLE}${REPO}/$1.git${NOFORMAT}"
  git clone "${START_URL}${REPO}/$1".git
  git -C "$1" config pull.rebase false
}

setup_colors
check_update "$0" "$@"
parse_params "$@"

REPO='flatfeestack'
PROJECTS='analysis-engine backend fastauth frontend payout'

for name in ${PROJECTS}; do
  if [ ! -d "$name" ]; then
    git_clone "$name" &
  else
    git -C "$name" pull &
  fi
done
wait

for name in ${PROJECTS}; do
  msg "${GREEN}[$(git -C "$name" symbolic-ref --short HEAD)]${NOFORMAT}-> $name"
done