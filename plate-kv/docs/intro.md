# Plate-KV HTTP API Documentation

## Overview

Plate-KV is a Redis-compatible HTTP API server that provides REST-like access to key-value data operations. Each `plateID` is an isolated namespace, so multiple applications or environments can share one server without colliding keys.

## Base URL

```text
%%URL%%/{plateID}/...
```

All endpoints require a `plateID` path parameter.

## Authentication

All endpoints require an `Authorization` header with your API key:

```text
Authorization: YOUR_API_KEY
```

See `authentication.md` for more details.

## Request Styles

Plate-KV supports two API styles:

- Wrapper endpoints are the preferred modern HTTP interface.
- Command endpoints remain available for backward compatibility and advanced Redis-compatible access.

### Wrapper Endpoints

Wrapper endpoints use descriptive URLs and typed JSON bodies.

Example:

```json
POST /{plateID}/hashes/set
{
  "key": "user:1",
  "value": {
    "name": "Ada",
    "age": 30
  }
}
```

### Command Endpoints

Command endpoints use a Redis-style request payload:

```json
{
  "command": "SET",
  "args": ["mykey", "hello world"]
}
```

## Response Format

Success responses use a consistent envelope:

```json
{
  "ok": true,
  "data": {
    "result": "OK"
  }
}
```

Some direct endpoints return typed payloads inside `data` instead of `data.result`, such as `/{plateID}/keys/{key}` and `/{plateID}/json/{key}`.

Error responses use:

```json
{
  "ok": false,
  "error": {
    "code": "invalid_request",
    "message": "human readable message"
  }
}
```

## API Categories

| Category | Preferred Prefix | Description |
| --- | --- | --- |
| Keys | `/keys/` | Key inspection, delete, rename, copy, expiry, scan |
| Strings | `/strings/` | String values, counters, ranges |
| Hashes | `/hashes/` | Field-value storage |
| Lists | `/lists/` | Ordered collections |
| Sets | `/sets/` | Unique member collections |
| Sorted Sets | `/zsets/` | Ranked collections |
| Streams | `/streams/` | Append-only event data |
| Bitmaps | `/bitmaps/` | Efficient bit storage |
| Geo | `/geo/` | Geospatial storage and search |
| JSON | `/json/` | Store and retrieve JSON values |
| Pub/Sub | `/pubsub/`, `/publish/`, `/subscribe/` | Publish and subscribe |
| Pipeline | `/pipeline` | Batch command execution |
| Transaction | `/transaction` | Atomic transaction execution |
| Info | `/info` | Plate statistics |

## Notes

- Pipelines and transactions are unchanged.
- Wrapper endpoints still execute through the command layer internally, so plate namespacing and command rewriting continue to apply.
- Keys or fields containing `/` should prefer body-based wrapper endpoints rather than path-based variants.
