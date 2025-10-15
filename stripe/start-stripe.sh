#!/usr/bin/env sh

set -euo pipefail

echo "Setup Stripe Listener"
#Since the .env is directly mounted, we need to work on a copy and overwrite the .env afterwards
#Trying directly with sed -i to edit in place wants to move the file, and the resulting error is:
#sed: can't move '/root/.config/stripe/.env' to '/root/.config/stripe/.env.bak': Resource busy
cp /root/.env /tmp/.env

#Ask user to go to website, wait at most 120sec, try to login/get secret
if timeout 120 stripe listen --print-secret 2>&1 | tee /tmp/stripe_output; then
  # First attempt succeeded, get the secret from the output
  STRIPE_SECRET_WEBHOOK=$(cat /tmp/stripe_output)
else
  # First attempt failed/timed out, verify if we're authenticated
  if ! stripe config --list > /dev/null 2>&1; then
    echo "Authentication failed"
    exit 1
  fi
  
  # We're authenticated, get the secret now
  echo "Getting webhook secret..."
  STRIPE_SECRET_WEBHOOK=$(stripe listen --print-secret)
fi

# Get stripe config once
STRIPE_CONFIG=$(stripe config --list)

# Update the env file with all values
sed -i "s/^STRIPE_SECRET_WEBHOOK=.*/STRIPE_SECRET_WEBHOOK=${STRIPE_SECRET_WEBHOOK}/" /tmp/.env
STRIPE_PUBLIC_API=$(echo "$STRIPE_CONFIG" | grep 'test_mode_pub_key' | awk -F '= ' '{print $2}' | tr -d \')
sed -i "s/^STRIPE_PUBLIC_API=.*/STRIPE_PUBLIC_API=${STRIPE_PUBLIC_API}/" /tmp/.env
STRIPE_SECRET_API=$(echo "$STRIPE_CONFIG" | grep 'test_mode_api_key' | awk -F '= ' '{print $2}' | tr -d \')
sed -i "s/^STRIPE_SECRET_API=.*/STRIPE_SECRET_API=${STRIPE_SECRET_API}/" /tmp/.env
cat /tmp/.env > /root/.env
