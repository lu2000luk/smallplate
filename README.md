# smallplate

A small firebase clone for having a lot of projects and not many servers.

## Structure

- plate-web: React dashboard for managing plates
- plate-manager: Central API and orchestration for plates
- plate-docs: Documentation for API and usage
- plate-vec: Vector service based on Chroma
- plate-kv: Key-value service based on Valkey/Redis
- plate-db: SQLite-based database
- plate-link: Shortlinks service
- helpers: CLI utils for development, testing and deployment

## Features in progress

- Client sockets for plate-kv (for pub/sub)
- Serverless functions
- Deployment docs