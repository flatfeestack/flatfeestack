FROM caddy:2-alpine

RUN apk add --update --no-cache curl

COPY ./Caddyfile /etc/caddy/Caddyfile

HEALTHCHECK --interval=30s --timeout=3s \
  CMD curl -f http://localhost:8080 || exit 1
