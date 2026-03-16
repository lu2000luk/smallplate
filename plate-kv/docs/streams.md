# Streams API

Streams are append-only logs with entry IDs and field-value payloads. They work well for events, feeds, message processing, and audit trails.

## Endpoints

| Method | Endpoint | Description |
| --- | --- | --- |
| POST | `/{plateID}/streams/add` | Add an entry |
| GET | `/{plateID}/streams/length/{key}` | Get stream length |
| GET | `/{plateID}/streams/range/{key}` | Forward range |
| GET | `/{plateID}/streams/reverse/{key}` | Reverse range |
| POST | `/{plateID}/streams/read` | `XREAD` |
| POST | `/{plateID}/streams/trim` | `XTRIM` |
| POST | `/{plateID}/streams/delete` | `XDEL` |
| GET | `/{plateID}/streams/info/{key}` | `XINFO` |
| POST | `/{plateID}/streams/groups/create` | Create a consumer group |
| POST | `/{plateID}/streams/groups/read` | `XREADGROUP` |
| POST | `/{plateID}/streams/groups/ack` | `XACK` |
| GET | `/{plateID}/streams/groups/pending/{key}/{group}` | `XPENDING` |
| POST | `/{plateID}/streams/groups/claim` | `XCLAIM` |
| POST | `/{plateID}/streams/groups/autoclaim` | `XAUTOCLAIM` |
| POST | `/{plateID}/streams/command` | Execute allowed stream commands |
| POST | `/{plateID}/streams/{key}/command` | Execute key-specific stream commands |

## Examples

Add an entry:

```json
POST /{plateID}/streams/add
{
  "key": "events",
  "values": {
    "type": "signup",
    "user_id": 123
  }
}
```

Read a range:

```text
GET /{plateID}/streams/range/events?start=-&end=+&count=10
```

Read from multiple streams:

```json
POST /{plateID}/streams/read
{
  "streams": [
    { "key": "events", "id": "$" },
    { "key": "audit", "id": "$" }
  ],
  "block_ms": 5000,
  "count": 25
}
```

Trim by max length:

```json
POST /{plateID}/streams/trim
{
  "key": "events",
  "strategy": "maxlen",
  "threshold": 1000,
  "approximate": true
}
```

Create a group:

```json
POST /{plateID}/streams/groups/create
{
  "key": "events",
  "group": "workers",
  "id": "$",
  "mkstream": true
}
```

Read as a consumer group:

```json
POST /{plateID}/streams/groups/read
{
  "group": "workers",
  "consumer": "worker-1",
  "streams": [
    { "key": "events", "id": ">" }
  ],
  "count": 10
}
```

## Command Compatibility

```json
POST /{plateID}/streams/command
{
  "command": "XADD",
  "args": ["events", "*", "type", "signup"]
}
```

## Command Endpoints

### `POST /{plateID}/streams/command`

Execute allowed stream commands across the plate.

**Allowed Commands:**

| Command | Description |
|---------|-------------|
| XADD | Add entry |
| XLEN | Get length |
| XRANGE | Forward range |
| XREVRANGE | Reverse range |
| XREAD | Read from streams |
| XTRIM | Trim stream |
| XDEL | Delete entries |
| XINFO | Get stream info |
| XGROUP | Manage groups |
| XREADGROUP | Group read |
| XACK | Acknowledge |
| XPENDING | Pending entries |
| XCLAIM | Claim messages |
| XAUTOCLAIM | Auto claim |

**Query Parameters for XRANGE/XREVRANGE:**
- `start` (optional): Start ID (default: `-` for XRANGE, `+` for XREVRANGE)
- `end` (optional): End ID (default: `+` for XRANGE, `-` for XREVRANGE)
- `count` (optional): Limit results

**Query Parameters for XINFO:**
- `section` (optional): `full`, `stream`, `groups`, `consumers`

**Query Parameters for XPENDING:**
- `start` (optional): Start ID
- `end` (optional): End ID
- `count` (optional): Number of entries
- `consumer` (optional): Filter by consumer

### `POST /{plateID}/streams/{key}/command`

Execute allowed commands on a specific stream key.

**Allowed Commands:** Same as above.

**Request:**

```json
POST /{plateID}/streams/events/command
{
  "command": "XLEN"
}
```
