#!/usr/bin/env bash

PROJECTS='db caddy anvil auth analyzer backend frontend stripe-webhook'

# Improved version based on https://betterdev.blog/minimal-safe-bash-script-template/
set -Eeuo pipefail
trap 'cleanup $?' SIGINT SIGTERM ERR EXIT

cleanup() {
  trap - SIGINT SIGTERM ERR EXIT
  # Add any cleanup tasks here
}

setup_colors() {
  if [[ -t 2 ]] && [[ -z "${NO_COLOR-}" ]] && [[ "${TERM-}" != "dumb" ]]; then
    NOFMT='\033[0m' RED='\033[0;31m' GREEN='\033[0;32m' YELLOW='\033[1;33m' BLUE='\033[0;34m'
  else
    NOFMT='' RED='' GREEN='' YELLOW='' BLUE=''
  fi
}

msg() {
  echo >&2 -e "${1-}"
}

msg_ok() {
  msg "${GREEN}${1-}${NOFMT}"
}

msg_warn() {
  msg "${YELLOW}WARN: ${1-}${NOFMT}"
}

msg_info() {
  msg "${BLUE}INFO: ${1-}${NOFMT}"
}

die() {
  msg "${RED}ERR: ${1-}${NOFMT}"
  exit "${2-1}"
}

usage() {
  cat <<EOF
Usage: $(basename "${BASH_SOURCE[0]}") [OPTIONS]
Build and run FlatFeeStack.

OPTIONS:
  -h, --help          Print this help and exit
  --no-stripe         Skip Stripe setup
  --clean             Remove all data (database, repos, chain, stripe, node_modules)
  --clean-db          Remove database only
EOF
  exit
}

check_envs() {
  local project="$1"
  local example_file="${2:-.example.env}"
  local target_file="${3:-.env}"
  
  local project_example="$project/$example_file"
  local project_target="$project/$target_file"
  
  # Skip if no example file exists
  [[ ! -f "$project_example" ]] && return 0
  
  # Create .env from example if missing
  if [[ ! -f "$project_target" ]]; then
    cp "$project_example" "$project_target"
    msg_ok "Created $project_target from $example_file"
    return 0
  fi
  
  # Check for missing keys
  local example_keys=$(grep -o '^[^#]*=' "$project_example" | sort)
  local target_keys=$(grep -o '^[^#]*=' "$project_target" | sort)
  
  if ! diff -q <(echo "$example_keys") <(echo "$target_keys") >/dev/null 2>&1; then
    local missing_keys=$(comm -23 <(echo "$example_keys") <(echo "$target_keys"))
    if [[ -n "$missing_keys" ]]; then
      msg_warn "$project_target is missing keys: $(echo "$missing_keys" | tr '\n' ' ')"
    fi
    
    local extra_keys=$(comm -13 <(echo "$example_keys") <(echo "$target_keys"))
    if [[ -n "$extra_keys" ]]; then
      msg_info "$project_target has extra keys: $(echo "$extra_keys" | tr '\n' ' ')"
    fi
  fi
}

parse_params() {
  # default values of variables set from params
  stripe=true
  internal="$PROJECTS"

  while :; do
    case "${1-}" in
    -h | --help) usage ;;
    --no-color) NO_COLOR=1 ;;
    --no-stripe) 
      internal="${internal//stripe-webhook/}"
      stripe=false
      ;;
    --clean) 
      sudo rm -rf .db/** .repos/** .chain/** .stripe/** frontend/node_modules/**
      msg_ok "Cleaned all data"
      ;;
    --clean-db) 
      sudo rm -rf .db/**
      msg_ok "Cleaned database"
      ;;
    -?*) die "Unknown option: $1" ;;
    *) break ;;
    esac
    shift
  done
}

setup_colors
parse_params "$@"

mkdir -p .db .repos .chain .stripe frontend/node_modules

check_envs "analyzer"
check_envs "backend"
check_envs "forum"
check_envs "auth"
check_envs "db"

if ! docker info >/dev/null 2>&1; then
  die "Docker is not running. Please start Docker and try again"
fi

export GID=$(id -g)

if [ "$stripe" = true ]; then
  msg "Setup Stripe"
  docker compose build stripe-setup
  if ! docker compose run --rm stripe-setup; then
    # this may happen on an expired, you get a 401, but we can remove the .stripe directory
    # then, it will try with login
    rm -rf .stripe
    if ! docker compose run --rm stripe-setup; then
      die "Stripe setup failed. Cannot continue without proper Stripe configuration"
    fi
  fi
  msg_ok "Stripe setup done"
else
  msg_info "Skip Stripe setup"
fi

# https://stackoverflow.com/questions/56844746/how-to-set-uid-and-gid-in-docker-compose
# https://hub.docker.com/_/postgres
msg_ok "Run: docker compose up --build --abort-on-container-exit ${internal}"
docker compose up --build --abort-on-container-exit ${internal}