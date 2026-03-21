# Integration examples

These snippets show how to mount redirects on your own domain (for example `/u/:id`) and resolve destination URLs through the JSON resolve endpoint:

`GET /{plateId}/resolve/{id}`

Each snippet follows this flow:

1. Receive a friendly route like `/u/:id/:tail*`.
2. Call plate-link resolve endpoint for your plate.
3. Read `data.destination` from JSON.
4. Redirect with HTTP 307.

Files:

- `nextjs.ts`
- `cloudflare-worker.js`
- `vite-middleware.ts`
- `nitro-server.ts`
- `go-http.go`
- `actix-web.rs`
- `rocket.rs`
