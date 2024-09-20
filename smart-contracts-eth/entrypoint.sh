#!/usr/bin/env sh

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

#start the node, after it is reachable, deploy contracts
npx hardhat node &
wait4ports -q -t 10 tcp://0.0.0.0:8545
npx hardhat run --network localhost scripts/deploy.ts
wait
