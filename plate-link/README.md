# plate-link

A link shortening service for smallplate.

## Features

- Connects to plate-manager over websocket using the same auth/check protocol as other services.
- Uses Valkey/Redis for storage.
- Authenticated management endpoints for creating/updating/deleting links.
- Public redirect endpoint: `/url/{id}` and `/url/{id}/{tail...}`.
- Public JSON resolver endpoint with CORS: `/{plateID}/resolve/{id}`.
- Dynamic template links with placeholders, expiry, max uses, metadata.
- Custom ID prefix support.
- Random ID generation starts at 5 chars and auto-increases after collision retries.

## Environment

- `DB_URL` Valkey/Redis URL (required)
- `SERVICE_ID` unique link service id (required)
- `SERVICE_KEY` manager shared secret (required)
- `MANAGER_URL` manager host:port or ws url (required)
- `PUBLIC_URL` optional public url advertised to manager
- `HTTP_ADDR` default `:3600`

## Endpoints

Authenticated:

- `POST /{plateID}/create`
- `POST /{plateID}/create/dynamic`
- `GET /{plateID}/links`
- `GET /{plateID}/links/{id}`
- `POST /{plateID}/links/{id}/update`
- `POST /{plateID}/links/{id}/metadata`
- `DELETE /{plateID}/links/{id}`

Public:

- `GET /url/{id}`
- `GET /url/{id}/{tail...}`
- `GET /{plateID}/resolve/{id}`
- `GET /{plateID}/resolve/{id}/{tail...}`

## Request examples

Create static:

```json
{
  "destination": "https://example.com/product/abc",
  "expires_at": 0,
  "max_uses": 100,
  "id_prefix": "prod_",
  "metadata": { "campaign": "spring" }
}
```

Create dynamic:

```json
{
  "template": "https://mycoolwebsite.com/product/{}?ref={}",
  "id_prefix": "dyn_"
}
```

Produces behavior like:

- `/url/dyn_xY19a/test/google` -> `https://mycoolwebsite.com/product/test?ref=google`

Named query placeholder template:

```json
{
  "template": "https://test.com/{product}"
}
```

Produces behavior like:

- `/url/aB9xQ?product=test` -> `https://test.com/test`

## ID policy

- User cannot specify full IDs.
- Optional `id_prefix` is allowed.
- Random suffix charset: `[a-zA-Z0-9]`.
- Start suffix length at 5.
- Retry up to 10 collisions for current length.
- If still colliding, increase length to 6 and retry (then 7, etc.).

## Examples

See `examples/` for integrations:

- Next.js
- Cloudflare Workers
- Vite middleware
- Nitro
- Go net/http
- actix-web
- rocket.rs
