FROM node:lts-alpine AS base
RUN apk update && apk add git
WORKDIR /app
ENV REACT_APP_ETH_RPC_URL=http://localhost:8545
# https://stackoverflow.com/questions/69692842/error-message-error0308010cdigital-envelope-routinesunsupported
RUN git clone https://github.com/xops/expedition.git && cd expedition && npm install && NODE_OPTIONS=--openssl-legacy-provider npm run build

FROM caddy:latest
WORKDIR /app
COPY --from=base /app/expedition/build ./build
ENTRYPOINT ["caddy", "file-server", "--listen", ":3000", "--root", "/app/build"]
