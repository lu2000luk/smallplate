# plate-vec

Vector API service for Smallplate.

`plate-vec` exposes a plate-scoped HTTP API and proxies vector/database operations to an internal Chroma server.

## What it provides

- Plate-scoped auth (`Authorization` header) via `plate-manager`
- Strict JSON decoding (`DisallowUnknownFields`) and `{ ok, data|error }` envelopes
- Chroma-backed databases, collections, and records
- Optional embedding generation with pinned providers:
  - OpenAI: `https://api.openai.com/v1/embeddings`
  - OpenRouter: `https://openrouter.ai/api/v1/embeddings`
- Local metadata persistence under `/data` for:
  - embedding profiles
  - plate-level defaults
  - collection embedding settings

## Environment

Required:

- `SERVICE_ID` (example: `vec_main`)
- `SERVICE_KEY`
- `MANAGER_URL` (example: `plate-manager:3200`)
- `CHROMA_URL` (example: `http://vec-chroma:8000`)

Optional:

- `PUBLIC_URL`
- `HTTP_ADDR` (default `:3300`)
- `DATA_DIR` (default `data`)

## Run locally

```bash
go mod tidy
go run .
```

## Docker

The repository compose files already include:

- `vec-chroma` (internal)
- `plate-vec` (public on `3300`)

Start stack:

```bash
docker compose up --build
```

## API shape

Base URL:

```text
[base-url]/[plateID]/...
```

Required auth header:

```text
Authorization: YOUR_API_KEY
```

Optional embedding key header for caller-billed embedding requests:

```text
X-Embedding-Api-Key: YOUR_PROVIDER_KEY
```

### Core endpoints

- `POST /{plateID}`
- `GET /{plateID}`
- `GET /{plateID}/info`
- `GET /{plateID}/limits`

### Databases

- `GET /{plateID}/databases`
- `POST /{plateID}/databases`
- `GET /{plateID}/databases/{database}`
- `DELETE /{plateID}/databases/{database}`
- `POST /{plateID}/databases/{database}/reset`

### Collections

- `GET /{plateID}/databases/{database}/collections`
- `GET /{plateID}/databases/{database}/collections/count`
- `POST /{plateID}/databases/{database}/collections`
- `GET /{plateID}/databases/{database}/collections/{collection}`
- `PATCH /{plateID}/databases/{database}/collections/{collection}`
- `DELETE /{plateID}/databases/{database}/collections/{collection}`
- `POST /{plateID}/databases/{database}/collections/{collection}/peek`

### Records

- `POST /{plateID}/databases/{database}/collections/{collection}/records/add`
- `POST /{plateID}/databases/{database}/collections/{collection}/records/update`
- `POST /{plateID}/databases/{database}/collections/{collection}/records/upsert`
- `POST /{plateID}/databases/{database}/collections/{collection}/records/get`
- `POST /{plateID}/databases/{database}/collections/{collection}/records/query`
- `POST /{plateID}/databases/{database}/collections/{collection}/records/delete`
- `GET /{plateID}/databases/{database}/collections/{collection}/records/count`

### Aliases and shortcuts

- Default DB aliases under `/{plateID}/collections/...`
- Document shortcuts:
  - `POST /{plateID}/collections/{collection}/documents/upsert`
  - `POST /{plateID}/collections/{collection}/documents/get`
  - `POST /{plateID}/collections/{collection}/documents/search`
  - `POST /{plateID}/collections/{collection}/documents/delete`
- Cross-collection search:
  - `POST /{plateID}/search`

### Embedding profile/default endpoints

- `GET /{plateID}/embedding-profiles`
- `POST /{plateID}/embedding-profiles`
- `PATCH /{plateID}/embedding-profiles/{profileID}`
- `DELETE /{plateID}/embedding-profiles/{profileID}`
- `GET /{plateID}/embedding-defaults`
- `PUT /{plateID}/embedding-defaults`

### Embedding utility/discovery

- `POST /{plateID}/embeddings/run`
- `GET /{plateID}/embedding-providers`
- `GET /{plateID}/embedding-models?provider=openai|openrouter`

### Import/export and operations

- `GET /{plateID}/export?database=default&format=jsonl|ndjson`
- `POST /{plateID}/import`
- `POST /{plateID}/databases/{database}/collections/{collection}/clone`
- `POST /{plateID}/databases/{database}/collections/{collection}/copy`
- `POST /{plateID}/databases/{database}/collections/{collection}/reindex`
- `GET /{plateID}/databases/{database}/collections/{collection}/stats`

## Embedding rules

- Only `provider`, `model`, and optional `dimensions` are accepted.
- `base_url`, `endpoint`, and custom provider endpoints are rejected.
- If request embedding config is absent, resolution order is:
  1. request override
  2. collection default profile
  3. plate default profile
  4. direct embeddings in request
  5. fail with `embedding_not_configured`

## Response envelope

Success:

```json
{
  "ok": true,
  "data": {}
}
```

Error:

```json
{
  "ok": false,
  "error": {
    "code": "invalid_request",
    "message": "human readable message"
  }
}
```

## Health

- `GET /health`
- `GET /healthz`
