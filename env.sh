#!/bin/bash -e

now=$(date +"%Y-%m-%d_%H-%M-%S")
tar cfz "env-${now}.tar.gz" \
 analysis-engine/.env \
 backend/.env \
 fastauth/.env \
 payout/.env \
 payout-nodejs/.env \
 .env