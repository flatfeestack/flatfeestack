#!/usr/bin/env bash

# based on https://betterdev.blog/minimal-safe-bash-script-template/
set -Eeuo pipefail
trap cleanup SIGINT SIGTERM ERR EXIT

cleanup() {
  trap - SIGINT SIGTERM ERR EXIT
  # script cleanup here
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

setup_colors

projects='analysis-engine backend fastauth frontend payout'

git pull &
for name in ${projects}; do
  [ ! -d "$name" ] && git clone git@github.com:flatfeestack/"$name".git;git -C "$name" config pull.rebase false;
  git -C "$name" pull &
done

wait

for name in ${projects}; do
  msg "${GREEN}[$(git -C "$name" symbolic-ref --short HEAD)]${NOFORMAT}-> $name"
done
