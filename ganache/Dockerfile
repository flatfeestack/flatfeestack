FROM node:lts-alpine
RUN apk update && apk add gcompat --no-cache
RUN npm install ganache --global
#TODO: harden it by not running as root
ENTRYPOINT ["ganache"]