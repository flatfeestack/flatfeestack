FROM node:15 AS base
WORKDIR /app
COPY package.json webpack.config.js tsconfig.json ./
COPY ./public/assets ./public/assets
COPY ./public/index.html ./public/
RUN npm install

FROM base as builder
COPY ./src ./src
RUN npm run build

FROM caddy:2
COPY Caddyfile /etc/caddy/Caddyfile
COPY --from=builder /app/public /var/www/html
