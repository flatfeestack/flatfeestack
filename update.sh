#!/bin/sh

git submodule update --init --recursive --remote -j 8
git submodule foreach 'git checkout master'
