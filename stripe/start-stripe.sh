#!/usr/bin/env sh

#Abort on error
set -euo pipefail

#Only do setup if this file is present, that is mounted in docker-compose. This means we want to
#setup the strip env vars for our backend. Once this is done, the data is valid for 90 days.
#after 90 days, you have to run 'docker compose up stripe-setup' again.
if [ -f /root/.env ]; then
  echo "Setup Stripe Listener"
  #Since the .env is directly mounted, we need to work on a copy and overwrite the .env afterwards
  #Trying directly with sed -i to edit in place wants to move the file, and the resulting error is:
  #sed: can't move '/root/.config/stripe/.env' to '/root/.config/stripe/.env.bak': Resource busy
  cp /root/.env /tmp/.env
  #Ugly hack: this triggers the login if not logged in yet, but exits with an error
  stripe listen --print-secret || true
  #Do it again, this time we are logged in
  STRIPE_SECRET_WEBHOOK=$(stripe listen --print-secret)
  sed -i "s/^STRIPE_SECRET_WEBHOOK=.*/STRIPE_SECRET_API=${STRIPE_SECRET_WEBHOOK}/" /tmp/.env
  STRIPE_PUBLIC_API=$(stripe config --list | grep 'test_mode_pub_key' | awk -F '= ' '{print $2}' | tr -d \')
  sed -i "s/^STRIPE_PUBLIC_API=.*/STRIPE_PUBLIC_API=${STRIPE_PUBLIC_API}/" /tmp/.env
  STRIPE_SECRET_API=$(stripe config --list | grep 'test_mode_api_key' | awk -F '= ' '{print $2}' | tr -d \')
  sed -i "s/^STRIPE_SECRET_API=.*/STRIPE_SECRET_API=${STRIPE_SECRET_API}/" /tmp/.env
  cat /tmp/.env > /root/.config/stripe/.env
else
  echo "Starting Stripe Listener"
  stripe listen --skip-verify --forward-to http://backend:9082
fi
