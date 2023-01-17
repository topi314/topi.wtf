FROM node:18-alpine as BUILD

WORKDIR /build

COPY package*.json ./

RUN npm ci

COPY . .

RUN npx nuxi generate

FROM node:18-alpine

COPY --from=BUILD /build/.output /app

EXPOSE 8080

ENV HOST=0.0.0.0
ENV PORT=8080
ENV NODE_ENV=production

ENTRYPOINT ["dumb-init", "node", "/app/server/index.mjs"]