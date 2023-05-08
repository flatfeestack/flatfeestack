#!/bin/bash -e

now=$(date +"%Y-%m-%d_%H-%M-%S")

find . \
    -maxdepth 2 \
    -type d ! \
    -executable \
    -prune \
    -o \
    -name ".env" \
    -type f \
    -printf "%P\n" | tar -czvf "flatfeestack-env-${now}.tar.gz" -T -
