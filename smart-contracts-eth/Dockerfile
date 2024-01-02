FROM node:20-alpine3.19
WORKDIR /app
RUN echo "hallo"
RUN apk add --update --no-cache wait4ports
COPY entrypoint.sh package.json hardhat.config.ts pnpm-lock.yaml tsconfig.json ./
RUN npm install -g pnpm && pnpm install --frozen-lockfile
COPY contracts ./contracts
COPY scripts ./scripts
#the compiled contracts are cached, so this is done only once
RUN npx hardhat compile
ENTRYPOINT ["sh", "entrypoint.sh"]
