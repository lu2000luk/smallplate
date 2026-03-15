# Pub/Sub API

The Pub/Sub API provides **publish/subscribe** messaging capabilities. This allows real-time communication between different parts of your application!

## Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/{plateID}/publish/{channel}` | Publish a message |
| GET | `/{plateID}/subscribe/{channel}` | Subscribe via SSE |
| GET | `/{plateID}/ws/subscribe/{channel}` | Subscribe via WebSocket |

---

## What is Pub/Sub?

Publish/Subscribe is a messaging pattern:
- **Publishers** send messages to channels
- **Subscribers** listen to channels and receive messages
- Publishers don't know who the subscribers are (decoupled!)

Use cases:
- Real-time notifications
- Chat systems
- Live updates
- Event-driven architecture

---

## Publish a Message

Send a message to a channel.

### POST /{plateID}/publish/{channel}

**Request:**
```json
{
  "message": { "text": "Hello, world!" }
}
```

**Response:**
```json
{
  "receivers": 2,
  "channel": "mychannel"
}
```
`receivers` = number of subscribers currently listening.

### JavaScript

```javascript
const response = await fetch("%%URL%%/my-plate/publish/mychannel", {
  method: "POST",
  headers: {
    "Authorization": "YOUR_API_KEY",
    "Content-Type": "application/json"
  },
  body: JSON.stringify({
    message: { text: "Hello, world!" }
  })
});
// result: { receivers: 2, channel: "mychannel" }
```

### cURL

```bash
curl -X POST "%%URL%%/my-plate/publish/alerts" \
  -H "Authorization: YOUR_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{"message":{"text":"Server restarted"}}'
```

### Different Message Types

```javascript
// Simple string
await fetch("%%URL%%/my-plate/publish/alerts", {
  method: "POST",
  headers: {
    "Authorization": "YOUR_API_KEY",
    "Content-Type": "application/json"
  },
  body: JSON.stringify({
    message: "Server restarted"
  })
});

// Complex object
await fetch("%%URL%%/my-plate/publish/user:123:events", {
  method: "POST",
  headers: {
    "Authorization": "YOUR_API_KEY",
    "Content-Type": "application/json"
  },
  body: JSON.stringify({
    message: {
      type: "login",
      userId: "123",
      timestamp: "2024-01-01T12:00:00Z"
    }
  })
});

// Array
await fetch("%%URL%%/my-plate/publish/tags", {
  method: "POST",
  headers: {
    "Authorization": "YOUR_API_KEY",
    "Content-Type": "application/json"
  },
  body: JSON.stringify({
    message: ["important", "urgent"]
  })
});
```

---

## Subscribe via SSE (Server-Sent Events)

Long-lived HTTP connection that streams messages in real-time.

### GET /{plateID}/subscribe/{channel}

**Query Parameters:**
- `pattern` (bool): Use pattern matching

**Response Format:**
```
event: message
data: {"channel":"mychannel","message":{"text":"Hello"}}
```

### JavaScript

```javascript
const eventSource = new EventSource("%%URL%%/my-plate/subscribe/mychannel", {
  headers: { "Authorization": "YOUR_API_KEY" }
});

eventSource.addEventListener("message", (event) => {
  const data = JSON.parse(event.data);
  console.log("Channel:", data.channel);
  console.log("Message:", data.message);
});

eventSource.onerror = () => {
  console.log("Connection closed");
};
```

### cURL

```bash
# This won't show real-time in terminal, but shows the endpoint structure
curl -N "%%URL%%/my-plate/subscribe/mychannel" \
  -H "Authorization: YOUR_API_KEY"
```

---

## Subscribe via WebSocket

Full-duplex real-time communication.

### GET /{plateID}/ws/subscribe/{channel}

**Query Parameters:**
- `pattern` (bool): Use pattern matching

### JavaScript (WebSocket)

```javascript
const ws = new WebSocket("%%URL%%/my-plate/ws/subscribe/mychannel");

ws.onopen = () => {
  console.log("Connected!");
};

ws.onmessage = (event) => {
  const data = JSON.parse(event.data);
  console.log("Channel:", data.channel);
  console.log("Message:", data.message);
};

ws.onclose = () => {
  console.log("Disconnected");
};
```

### cURL (using wscat or similar)

```bash
# If you have wscat installed
wscat -c "ws://%%URL%%/my-plate/ws/subscribe/mychannel" \
  -H "Authorization: YOUR_API_KEY"
```

---

## Pattern Subscribe

Subscribe to multiple channels matching a pattern using wildcards.

### SSE with Pattern

```javascript
const eventSource = new EventSource(
  "%%URL%%/my-plate/subscribe/news:*?pattern=true",
  { headers: { "Authorization": "YOUR_API_KEY" } }
);
```

Matches: `news:sports`, `news:tech`, `news:finance`, etc.

### WebSocket with Pattern

```javascript
const ws = new WebSocket(
  "%%URL%%/my-plate/ws/subscribe/user:*?pattern=true",
  [],
  { "Authorization": "YOUR_API_KEY" }
);
```

Matches: `user:123`, `user:456`, etc.

### cURL

```bash
# Pattern matching via SSE
curl -N "%%URL%%/my-plate/subscribe/news:*?pattern=true" \
  -H "Authorization: YOUR_API_KEY"
```

---

## Message Format

Messages received via SSE or WebSocket:

```json
{
  "channel": "mychannel",
  "message": { "text": "Hello" }
}
```

If the published message is valid JSON, it's parsed. Otherwise, delivered as a string.

---

## Keep-Alive

SSE connections send keepalive messages every 15 seconds:

```
: keepalive
```

This prevents proxies from closing the connection.

---

## Error Responses

```json
{
  "error": "missing_authorization",
  "message": "authorization header is required"
}
```

```json
{
  "error": "streaming_unsupported",
  "message": "response writer does not support streaming"
}
```

---

## Use Cases

### Real-time Notifications

```javascript
// Publisher: Send notification
await fetch("%%URL%%/my-plate/publish/notifications", {
  method: "POST",
  headers: {
    "Authorization": "YOUR_API_KEY",
    "Content-Type": "application/json"
  },
  body: JSON.stringify({
    message: { type: "new_message", from: "user123" }
  })
});

// Subscriber: Listen for notifications
const eventSource = new EventSource(
  "%%URL%%/my-plate/subscribe/notifications",
  { headers: { "Authorization": "YOUR_API_KEY" } }
);
```

### Live Updates

```javascript
// Publish stock price updates
await fetch("%%URL%%/my-plate/publish/stocks/AAPL", {
  method: "POST",
  headers: {
    "Authorization": "YOUR_API_KEY",
    "Content-Type": "application/json"
  },
  body: JSON.stringify({
    message: { price: 150.25, time: "2024-01-01T12:00:00Z" }
  })
});

// Subscribe to stock updates
const eventSource = new EventSource(
  "%%URL%%/my-plate/subscribe/stocks:*?pattern=true",
  { headers: { "Authorization": "YOUR_API_KEY" } }
);
```