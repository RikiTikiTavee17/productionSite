FROM node:18-alpine

WORKDIR /app
COPY package.json package-lock.json ./
RUN npm install
RUN npm install webpack webpack-cli --save-dev

COPY . .

RUN  npx webpack --config webpack.config.js

EXPOSE 3000
CMD ["npx", "serve", "dist", "-l", "3000"]