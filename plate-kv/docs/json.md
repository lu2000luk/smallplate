# JSON API

JSON endpoints make it easy to store structured JSON without manually serializing and deserializing at the client layer.

## Endpoints

| Method | Endpoint | Description |
| --- | --- | --- |
| GET | `/{plateID}/json/{key}` | Get a JSON value |
| POST | `/{plateID}/json/{key}` | Store a JSON value |
| DELETE | `/{plateID}/json/{key}` | Delete a JSON value |

## Examples

Store JSON:

```json
POST /{plateID}/json/user:1
{
  "value": {
    "name": "Ada",
    "age": 30,
    "active": true
  },
  "ttl_ms": 3600000
}
```

Read JSON:

```text
GET /{plateID}/json/user:1
```

Delete JSON:

```text
DELETE /{plateID}/json/user:1
```

## Notes

- Values are stored as JSON-encoded strings internally.
- `GET` returns the decoded JSON structure inside the standard envelope.
