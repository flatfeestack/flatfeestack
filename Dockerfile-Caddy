FROM node:12-alpine AS builder
WORKDIR /app
COPY ./frontend/src ./src
COPY ./frontend/public ./public
COPY ./frontend/package.json ./frontend/webpack.config.js ./frontend/tailwind.config.js ./frontend/tsconfig.json ./
RUN npm install
RUN npm run build


FROM caddy:2.0.0
COPY Caddyfile /etc/caddy/Caddyfile
COPY --from=builder /app/public /var/www/html
