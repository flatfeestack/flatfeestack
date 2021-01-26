#!/bin/sh
set -e

if [[ "$ENV" = "local" ]]; then
  LD_PRELOAD=/usr/local/lib/faketime/libfaketime.so.1 FAKETIME_NO_CACHE=1 exec ./backend
else
  exec ./backend
fi
