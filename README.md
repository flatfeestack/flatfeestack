# Flatfeestack Installation
This repo combines all Flatfeestack packages using `docker-compose`.

## Build / Start / Stop

```shell script
git clone --recurse-submodules https://github.com/flatfeestack/flatfeestack.git
cd flatfeestack
docker-compose up -d --build
```

For ubuntu, install:

```shell script
sudo apt install net-tools
```

if you want to stop, and clean everything up:

```shell script
docker-compose down -v
```

## Networking

This repo includes a caddy server to create reverse proxies to the different packages:

**/** --> Frontend

**/auth/*** --> Authentication Service

**/analysis/*** --> Analysis Engine

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

