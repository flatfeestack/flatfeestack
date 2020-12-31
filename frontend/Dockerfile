FROM node:14-alpine AS builder
RUN addgroup -S appgroup && adduser -S appuser -G appgroup
RUN apk add make gcc musl-dev
WORKDIR /app
COPY package.json webpack.config.js tailwind.config.js tsconfig.json ./
COPY ./public/assets ./public/assets
COPY ./public/favicon.png ./public/global.css ./public/index.html ./public/
RUN chown -R appuser:appgroup /app
USER appuser
RUN npm install
COPY --chown=appuser:appgroup ./src ./src
RUN npm run build

FROM caddy:2-alpine
COPY Caddyfile /etc/caddy/Caddyfile
COPY --from=builder /app/public /var/www/html
