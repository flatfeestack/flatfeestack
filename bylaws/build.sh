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

docker build . -t bylaws:latest
docker run --rm --entrypoint cat bylaws bylaws.html > bylaws.html
hash=$(sha256sum bylaws.html | awk '{print $1}')
mv bylaws.html bylaws-$(date -r bylaws.md +'%Y-%m-%d')_${hash}.html