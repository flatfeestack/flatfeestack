#!/bin/sh -e

git submodule update --recursive --remote
git submodule foreach 'git pull origin main'
