FROM node:15 AS base
WORKDIR /app
COPY package.json index.js ./
RUN npm install

FROM base as builder
#COPY ./src ./src
CMD ["npm", "run", "prod"]