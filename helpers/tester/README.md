# tester

To install dependencies:

```bash
bun install
```

To run:

```bash
bun run index.ts
```

## Recommended usage

Run against explicit service URLs and enable debug output:

```bash
bun run index.ts \
  -p <plate-id> \
  -k <api-key> \
  --db <db-url> \
  --kv <kv-url> \
  --vec <vec-url> \
  --link <link-url> \
  --debug \
  --non-interactive \
  --no-open
```

## Debugging flags

- `--debug` prints failure hints, reproducible cURL commands, and related docs paths.
- Exact request/response details are printed (headers + full body + resolved URL).
- `--services db,kv,vec,link` limits which services run.
- `--timeout-ms 30000` increases per-request timeout.
- `--report-dir ./reports` writes reports to a custom directory.
- `--non-interactive` skips prompts and runs automatically.
- `--no-open` prevents opening an editor after report generation.

## OpenRouter embedding test category

To run the extra OpenRouter embedding test category (under Vec), provide both:

- `--openrouter-key <your-openrouter-api-key>`
- `--openrouter-model <embedding-model-id>`

Example:

```bash
bun run index.ts \
  -p <plate-id> \
  -k <api-key> \
  --vec <vec-url> \
  --services vec \
  --openrouter-key <key> \
  --openrouter-model openai/text-embedding-3-small \
  --debug -y --no-open
```

## Service docs

- DB: `/docs/db/`
- KV: `/docs/kv/`
- Vec: `/docs/vec/`
- Link: `/docs/link/`

This project was created using `bun init` in bun v1.3.11. [Bun](https://bun.com) is a fast all-in-one JavaScript runtime.
