#!/bin/sh -e

git submodule update --init --recursive --remote -j 16
#git submodule foreach 'git checkout master'
git submodule foreach 'git config pull.rebase false'
