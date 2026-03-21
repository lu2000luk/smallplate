FROM oven/bun:1.2.23-alpine AS runner
WORKDIR /app

ENV NODE_ENV=production
ENV PORT=3200

COPY package.json bun.lock ./
RUN bun install --frozen-lockfile --production

COPY . .

EXPOSE 3200

CMD ["bun", "run", "index.ts"]
