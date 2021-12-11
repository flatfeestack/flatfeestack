#!/bin/sh 

command -v semver >/dev/null || {
  echo "Command 'semver' not found in \$PATH. Please, first install it." >&2
  exit 1
}

#git pull --recurse-submodules -j 16
##pulls to the latest version
#git submodule update --recursive --remote
##or
git fetch
#git submodule update --recursive --remote
git submodule foreach --recursive git pull origin main
git add analysis-engine backend fastauth frontend payout payout-nodejs search-proj
git commit -m "update to latest"
git push --recurse-submodules=on-demand
echo "get latest tag"
CURRENT=`git tag --sort=creatordate | tail -1`
git tag "`semver $CURRENT -i patch`"
git push origin `semver $CURRENT -i patch`
