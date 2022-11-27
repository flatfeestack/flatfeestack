#!/bin/bash -e

now=$(date +"%Y-%m-%d_%H-%M-%S")
tar cfz "flatfeestack-env-${now}.tar.gz" \
 analysis-engine/.env \
 backend/.env \
 db/.env \
 fastauth/.env \
 payout/.env \
 .env