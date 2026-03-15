# Pub/Sub API

Pub/Sub provides real-time message fan-out. Publishers send messages to a channel, and subscribers receive them over Server-Sent Events or WebSockets.

## Endpoints

| Method | Endpoint | Description |
| --- | --- | --- |
| POST | `/{plateID}/pubsub/{channel}/publish` | Publish a message |
| GET | `/{plateID}/pubsub/{channel}/subscribe` | Subscribe with SSE |
| GET | `/{plateID}/pubsub/{channel}/ws` | Subscribe with WebSocket |
| POST | `/{plateID}/publish/{channel}` | Legacy publish alias |
| GET | `/{plateID}/subscribe/{channel}` | Legacy SSE alias |
| GET | `/{plateID}/ws/subscribe/{channel}` | Legacy WebSocket alias |

## Publish Examples

Simple string:

```json
POST /{plateID}/pubsub/alerts/publish
{
  "message": "Server restarted"
}
```

Structured object:

```json
POST /{plateID}/pubsub/user:123:events/publish
{
  "message": {
    "type": "login",
    "userId": "123",
    "timestamp": "2026-03-15T12:00:00Z"
  }
}
```

Array payload:

```json
POST /{plateID}/pubsub/tags/publish
{
  "message": ["important", "urgent"]
}
```

## Subscribe Examples

SSE:

```text
GET /{plateID}/pubsub/alerts/subscribe
```

Pattern SSE:

```text
GET /{plateID}/pubsub/user:*/subscribe?pattern=true
```

WebSocket:

```text
GET /{plateID}/pubsub/alerts/ws
```

## Notes

- Published payloads are JSON-encoded.
- Subscription payloads decode JSON when possible and fall back to plain strings.
- Legacy publish/subscribe routes are still supported.
