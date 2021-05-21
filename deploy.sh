#!/bin/sh -e

#git pull --recurse-submodules -j 16
##pulls to the latest version
#git submodule update --recursive --remote
##or
git submodule foreach 'git pull'
