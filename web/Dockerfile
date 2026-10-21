FROM node:24.22.1-bookworm-slim AS build

RUN corepack enable && corepack prepare pnpm@10.28.2 --activate
WORKDIR /app
COPY package.json pnpm-lock.yaml ./
COPY vendor ./vendor
RUN pnpm install --frozen-lockfile --ignore-scripts

COPY . .
ENV NODE_OPTIONS=--max-old-space-size=4096
RUN pnpm build

FROM node:24.22.1-bookworm-slim

WORKDIR /app
COPY --from=build /app/.output ./.output

ENV HOST=0.0.0.0
ENV PORT=3007

EXPOSE 3007
CMD ["node", ".output/server/index.mjs"]
