# Flatfeestack Installation
This repo combines all Flatfeestack packages using `docker-compose`.

## Build / Start / Stop (local development)

```shell script
./update.sh

# Create example .env files
cp analyzer/.example.env analyzer/.env
cp backend/example.env backend/.env
cp fastauth/example.env fastauth/.env
cp payout/.example.env payout/.env
echo "HOST=http://localhost:8080" >> caddy/.env

mkdir db
echo "POSTGRES_PASSWORD=password" > db/.env
echo "POSTGRES_USER=postgres" >> db/.env
echo "POSTGRES_DB=flatfeestack" >> db/.env

# Build and run FlatFeeStack
docker compose build
docker compose up -d db
docker compose up
```

To register a user:

1. Open `localhost:8080` and register a new user.
2. Connect to the database: `docker exec -it flatfeestack-db-1 psql -U postgres`
3. Switch to the FlatFeeStack database: `\c flatfeestack`
4. Show content of the `auth` table: `TABLE auth;`.
5. Get the token for the user you just registered.
6. Exit the PSQL session and confirm the user registration: `curl "http://localhost:9081/confirm/signup/{email}/{token}"`

If you want to stop, and clean everything up:

```shell script
docker-compose down -v
```

## Networking

This repo includes a caddy server to create reverse proxies to the different packages:

**/** --> Frontend

**/auth/*** --> Authentication Service

**/analyzer/*** --> Analysis Engine

## Env

Sample .env

```
POSTGRES_PASSWORD=password
POSTGRES_USER=postgres
POSTGRES_DB=flatfeestack
```

## Tezos sandbox

Get wallet information on private blockchain
```
docker exec -it flatfeestack-flextesa flobox info
```

