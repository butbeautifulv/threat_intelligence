# syntax=docker/dockerfile:1
FROM node:20-bookworm AS deps
WORKDIR /app
COPY panel/package.json ./
RUN npm install

FROM deps AS build
COPY panel/ ./
RUN npm run build

FROM node:20-bookworm-slim AS runner
WORKDIR /app
ENV NODE_ENV=production
ENV PORT=8088
COPY --from=build /app/.next/standalone ./
COPY --from=build /app/.next/static ./.next/static
COPY --from=build /app/public ./public
EXPOSE 8088
USER node
CMD ["node", "server.js"]
