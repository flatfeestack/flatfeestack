#!/bin/sh -e

git submodule foreach 'git pull origin main'
git submodule foreach --recursive 'git checkout main'
