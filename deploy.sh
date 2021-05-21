#!/bin/sh -e

#git pull --recurse-submodules -j 16
##pulls to the latest version
#git submodule update --recursive --remote
##or
git submodule foreach 'git pull'
git add analysis-engine backend fastauth frontend payout search-proj
git commit -m "update to latest"
git push --recurse-submodules=on-demand
