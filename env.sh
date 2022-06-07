#!/bin/bash -e

now=$(date +"%Y-%m-%d_%H-%M-%S")
tar cfz "flatfeestack-env-${now}.tar.gz" \
 analysis-engine/.env \
 backend/.env \
 fastauth/.env \
 payout/.env \
 Caddyfile \
 .db_pw.txt \
 .env