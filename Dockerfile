FROM node:18-alpine as BUILD

WORKDIR /build

COPY package*.json ./

RUN npm ci

COPY . .

RUN npm run build

FROM node:18-alpine

RUN apk add --no-cache dumb-init

COPY --from=BUILD /build/.output /app

EXPOSE 8080

ENV HOST=0.0.0.0
ENV PORT=8080
ENV NODE_ENV=production

ENTRYPOINT ["/usr/bin/dumb-init", "--"]

CMD ["node", "/app/server/index.mjs"]