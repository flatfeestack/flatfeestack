FROM node:20-alpine AS base
WORKDIR /app
RUN apk --no-cache add curl
RUN curl -fsSL "https://github.com/pnpm/pnpm/releases/latest/download/pnpm-linuxstatic-x64" -o /bin/pnpm; chmod +x /bin/pnpm;
COPY package.json pnpm-lock.yaml ./
RUN pnpm install
COPY tsconfig.json tsconfig.node.json vite.config.ts svelte.config.js ./
COPY ./public ./public
COPY ./src ./src
RUN pnpm build

FROM caddy:2-alpine

RUN apk add --update --no-cache curl

COPY Caddyfile /etc/caddy/Caddyfile
COPY --from=base /app/dist/client /var/www/html

HEALTHCHECK --interval=30s --timeout=3s \
  CMD curl -f http://localhost:9085 || exit 1
