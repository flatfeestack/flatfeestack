FROM node:24-alpine AS base
RUN apk add --no-cache libc6-compat brotli gzip zstd parallel
WORKDIR /app
RUN npm install -g pnpm
COPY package.json pnpm-lock.yaml tsconfig.json ./
RUN pnpm install
COPY src public ./
RUN pnpm build

FROM caddy:2-alpine
COPY Caddyfile /etc/caddy/Caddyfile
COPY --from=base /app/dist/ /var/www/html