FROM oven/bun:1.2.23-alpine AS deps
WORKDIR /app

COPY package.json bun.lock ./
RUN bun install --frozen-lockfile --linker=hoisted

FROM oven/bun:1.2.23-alpine AS builder
WORKDIR /app

ENV NEXT_TELEMETRY_DISABLED=1
ARG NEXT_PUBLIC_MANAGER_URL
ARG NEXT_PUBLIC_DOCS_URL
ENV NEXT_PUBLIC_MANAGER_URL=${NEXT_PUBLIC_MANAGER_URL}
ENV NEXT_PUBLIC_DOCS_URL=${NEXT_PUBLIC_DOCS_URL}

COPY --from=deps /app/node_modules ./node_modules
COPY . .
RUN bun run build

FROM oven/bun:1.2.23-alpine AS runner
WORKDIR /app

ENV NODE_ENV=production
ENV NEXT_TELEMETRY_DISABLED=1
ENV PORT=3000
ARG NEXT_PUBLIC_MANAGER_URL
ARG NEXT_PUBLIC_DOCS_URL
ENV NEXT_PUBLIC_MANAGER_URL=${NEXT_PUBLIC_MANAGER_URL}
ENV NEXT_PUBLIC_DOCS_URL=${NEXT_PUBLIC_DOCS_URL}

COPY package.json bun.lock ./
RUN bun install --frozen-lockfile --production --linker=hoisted

COPY --from=builder /app/.next ./.next
COPY --from=builder /app/public ./public
COPY --from=builder /app/next.config.ts ./next.config.ts

EXPOSE 3000

CMD ["bun", "run", "start", "--", "-p", "3000", "-H", "0.0.0.0"]
