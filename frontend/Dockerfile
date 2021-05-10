FROM node:15 AS base
WORKDIR /app
COPY package.json tsconfig.json rollup-prod.config.js ssr.js ./
COPY ./public/assets ./public/assets
RUN npm install

FROM base as builder
COPY ./src ./src
RUN npm run prod

FROM caddy:2
COPY Caddyfile /etc/caddy/Caddyfile
COPY --from=builder /app/public /var/www/html
