FROM node:12-alpine

WORKDIR /usr/src/app

COPY webpack.config.js ./
COPY package*.json ./

RUN npm install

COPY . .

RUN npm run build

EXPOSE 3000

ENV HOST=0.0.0.0


CMD [ "npm", "start" ]
