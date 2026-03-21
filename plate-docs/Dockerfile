FROM oven/bun:1.2.23-alpine AS deps
WORKDIR /app

COPY package.json bun.lock ./
RUN bun install --frozen-lockfile --ignore-scripts

FROM oven/bun:1.2.23-alpine AS builder
WORKDIR /app

ENV NEXT_TELEMETRY_DISABLED=1

COPY --from=deps /app/node_modules ./node_modules
COPY . .
RUN bun run build

FROM oven/bun:1.2.23-alpine AS runner
WORKDIR /app

ENV NODE_ENV=production
ENV NEXT_TELEMETRY_DISABLED=1
ENV PORT=3001

COPY package.json bun.lock ./
RUN bun install --frozen-lockfile --production --ignore-scripts

COPY --from=builder /app/.next ./.next
COPY --from=builder /app/content ./content
COPY --from=builder /app/.source ./.source
COPY --from=builder /app/source.config.ts ./source.config.ts
COPY --from=builder /app/next.config.mjs ./next.config.mjs

EXPOSE 3001

CMD ["bun", "run", "start", "--", "-p", "3001", "-H", "0.0.0.0"]
