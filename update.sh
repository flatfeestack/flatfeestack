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

setup_colors
check_update "$0" "$@"

REPO='flatfeestack'
PROJECTS='analysis-engine backend fastauth frontend payout'

for name in ${PROJECTS}; do
  if [ ! -d "$name" ]; then
    git clone git@github.com:"$REPO"/"$name".git
    git -C "$name" config pull.rebase false;
  else
    git -C "$name" pull &
  fi
done
wait

#landing page
if [ ! -d "$name" ]; then
  git -C "frontend" clone git@github.com:flatfeestack/landing-page.git
  git -C "frontend/landing-page" config pull.rebase false
else
  git -C "frontend/landing-page" pull &
fi
wait

for name in ${PROJECTS}; do
  msg "${GREEN}[$(git -C "$name" symbolic-ref --short HEAD)]${NOFORMAT}-> $name"
done
#landing page
msg "${GREEN}[$(git -C "frontend/landing-page" symbolic-ref --short HEAD)]${NOFORMAT}-> frontend/landing-page"