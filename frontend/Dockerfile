FROM node:16-alpine AS base
WORKDIR /app
COPY package.json tsconfig.json rollup-prod.config.js generate-index.js ./
RUN npm install

FROM base as builder
COPY ./src ./src
COPY ./landing-page ./landing-page
RUN npm run prod

FROM caddy:2-alpine
COPY Caddyfile /etc/caddy/Caddyfile
COPY --from=builder /app/public /var/www/html
