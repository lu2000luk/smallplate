# Plate-KV HTTP API Documentation

## Overview

Plate-KV is a Redis-compatible HTTP API server that provides RESTful access to key-value data operations. Each "plate" represents an isolated namespace, allowing multiple independent datasets within a single server.

## Base URL

```
%%URL%%/{plateID}/...
```

All endpoints require a `plateID` path parameter that identifies the data partition.

## Authentication

All endpoints require authentication via the `Authorization` header with your API key:

```
Authorization: YOUR_API_KEY
```

See [authentication.md](authentication.md) for details.

## Request/Response Format

### Command Endpoints

Most data operations use a command-based pattern:

**Request:**
```json
{
  "command": "COMMAND_NAME",
  "args": ["arg1", "arg2", "..."]
}
```

**Response:**
```json
{
  "result": <value>
}
```

### JSON Endpoints

Some endpoints use direct REST patterns with JSON bodies:

**Request:**
```json
{
  "key": "value",
  "additional": "fields"
}
```

## API Categories

| Category | Endpoint Prefix | Description |
|----------|-----------------|-------------|
| Keys | `/keys/` | Key management, scanning, TTL |
| Strings | `/strings/` | String operations |
| Hashes | `/hashes/` | Hash/field-value operations |
| Lists | `/lists/` | List operations |
| Sets | `/sets/` | Set operations |
| Sorted Sets | `/zsets/` | Sorted set operations |
| Streams | `/streams/` | Stream operations |
| Bitmaps | `/bitmaps/` | Bitmap operations |
| Geo | `/geo/` | Geospatial operations |
| Pub/Sub | `/publish/`, `/subscribe/` | Publish/Subscribe messaging |
| Pipeline | `/pipeline` | Batch command execution |
| Transaction | `/transaction` | Atomic transaction execution |
| JSON | `/json/` | JSON storage |
| Info | `/info` | Plate statistics |

## Error Responses

Errors return appropriate HTTP status codes with JSON bodies:

```json
{
  "error": "error_code",
  "message": "human readable message"
}
```

Common status codes:
- `200 OK` - Success
- `400 Bad Request` - Invalid request format or parameters
- `401 Unauthorized` - Missing or invalid authentication
- `404 Not Found` - Key does not exist
- `500 Internal Server Error` - Server-side errors

## Rate Limiting

No built-in rate limiting. Implement at proxy level if needed.