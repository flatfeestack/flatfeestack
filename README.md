# Flatfeestack Docker
This repo combines all Flatfeestack packages using `docker-compose`.

## Build / Start
`docker-compose up -d --build`

## Networking

This repo includes a caddy server to create reverse proxies to the different packages:

**/** --> Frontend

**/auth/*** --> Authentication Service

**/analysis/*** --> Analysis Engine


