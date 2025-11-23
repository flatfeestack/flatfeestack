FROM node:24-alpine AS base
RUN apk add --no-cache pnpm
WORKDIR /app
COPY package.json ./
COPY artifacts/contracts ./artifacts/contracts
RUN pnpm install
ENTRYPOINT ["pnpm", "run", "dev"]