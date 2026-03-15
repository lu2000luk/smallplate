# Keys API

The Keys API provides operations for key management, inspection, scanning, and deletion.

## Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/{plateID}/keys/{key}` | Get key metadata and value |
| POST | `/{plateID}/keys/command` | Execute key commands |
| POST | `/{plateID}/keys/{key}/command` | Execute key-specific commands |
| GET | `/{plateID}/scan` | Scan keys in a plate |
| POST | `/{plateID}/scan/hashes/{key}` | Scan hash fields |
| POST | `/{plateID}/scan/sets/{key}` | Scan set members |
| POST | `/{plateID}/scan/zsets/{key}` | Scan sorted set members |
| DELETE | `/{plateID}/keys/{pattern}` | Delete keys by pattern |

---

## Understanding Keys

Keys are the fundamental identifier in key-value storage. Each key:

- Has a name (string)
- Holds a value of a specific type (string, list, hash, set, etc.)
- Can have an optional TTL (expiration time)
- Is unique within its plate

---

## GET /{plateID}/keys/{key}

Get metadata and value for a specific key.

**Response:**
```json
{
  "key": "mykey",
  "exists": true,
  "type": "string",
  "ttl_ms": 3600000,
  "value": "myvalue"
}
```

The response includes:
- `key` - the key name
- `exists` - whether the key exists
- `type` - data type (string, list, set, zset, hash, stream)
- `ttl_ms` - time until expiration in milliseconds (0 means no expiry)
- `value` - the actual value (only for string type)

### JavaScript

```javascript
const response = await fetch("%%URL%%/my-plate/keys/mykey", {
  headers: { "Authorization": "YOUR_API_KEY" }
});
const data = await response.json();
console.log(data);
// { key: "mykey", exists: true, type: "string", value: "myvalue" }
```

### cURL

```bash
curl -X GET "%%URL%%/my-plate/keys/mykey" \
  -H "Authorization: YOUR_API_KEY"
```

---

## POST /{plateID}/keys/command

Execute arbitrary key commands for any key.

**Request:**
```json
{
  "command": "SET",
  "args": ["mykey", "myvalue", "EX", "3600"]
}
```

**Response:**
```json
{
  "result": "OK"
}
```

### JavaScript

```javascript
const response = await fetch("%%URL%%/my-plate/keys/command", {
  method: "POST",
  headers: {
    "Authorization": "YOUR_API_KEY",
    "Content-Type": "application/json"
  },
  body: JSON.stringify({
    command: "SET",
    args: ["mykey", "myvalue", "EX", "3600"]
  })
});
```

### cURL

```bash
curl -X POST "%%URL%%/my-plate/keys/command" \
  -H "Authorization: YOUR_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{"command":"SET","args":["mykey","myvalue","EX","3600"]}'
```

---

## Common Key Commands

### DEL - Delete Keys

Remove one or more keys permanently.

```javascript
await fetch("%%URL%%/my-plate/keys/command", {
  method: "POST",
  headers: {
    "Authorization": "YOUR_API_KEY",
    "Content-Type": "application/json"
  },
  body: JSON.stringify({
    command: "DEL",
    args: ["key1", "key2"]
  })
});
// result: 2 (number of keys deleted)
```

```bash
curl -X POST "%%URL%%/my-plate/keys/command" \
  -H "Authorization: YOUR_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{"command":"DEL","args":["key1","key2"]}'
```

### EXISTS - Check if Keys Exist

Check if one or more keys exist.

```javascript
await fetch("%%URL%%/my-plate/keys/command", {
  method: "POST",
  headers: {
    "Authorization": "YOUR_API_KEY",
    "Content-Type": "application/json"
  },
  body: JSON.stringify({
    command: "EXISTS",
    args: ["key1", "key2"]
  })
});
// result: 1 (number of keys that exist)
```

### TYPE - Get Key Type

Get the data type of a key.

```javascript
await fetch("%%URL%%/my-plate/keys/command", {
  method: "POST",
  headers: {
    "Authorization": "YOUR_API_KEY",
    "Content-Type": "application/json"
  },
  body: JSON.stringify({
    command: "TYPE",
    args: ["mykey"]
  })
});
// result: "string" (or "list", "set", "zset", "hash", "stream")
```

---

## Setting Expiration (TTL)

Keys can automatically expire after a certain time.

### EXPIRE - Set Expiration in Seconds

```javascript
// Set key to expire in 1 hour
await fetch("%%URL%%/my-plate/keys/command", {
  method: "POST",
  headers: {
    "Authorization": "YOUR_API_KEY",
    "Content-Type": "application/json"
  },
  body: JSON.stringify({
    command: "EXPIRE",
    args: ["mykey", "3600"]
  })
});
// result: 1 (success)
```

### PEXPIRE - Set Expiration in Milliseconds

```bash
# Expire in 30 minutes (1800000 ms)
curl -X POST "%%URL%%/my-plate/keys/command" \
  -H "Authorization: YOUR_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{"command":"PEXPIRE","args":["mykey","1800000"]}'
```

### TTL - Get Time Until Expiration (seconds)

```javascript
await fetch("%%URL%%/my-plate/keys/command", {
  method: "POST",
  headers: {
    "Authorization": "YOUR_API_KEY",
    "Content-Type": "application/json"
  },
  body: JSON.stringify({
    command: "TTL",
    args: ["mykey"]
  })
});
// result: 3500 (seconds remaining, -1 means no expiry, -2 means key doesn't exist)
```

### PTTL - Get Time Until Expiration (milliseconds)

Same as TTL but returns milliseconds.

### PERSIST - Remove Expiration

Remove the expiration from a key (make it permanent).

```javascript
await fetch("%%URL%%/my-plate/keys/command", {
  method: "POST",
  headers: {
    "Authorization": "YOUR_API_KEY",
    "Content-Type": "application/json"
  },
  body: JSON.stringify({
    command: "PERSIST",
    args: ["mykey"]
  })
});
// result: 1 (success)
```

---

## RENAME - Rename a Key

Change the name of a key.

```javascript
await fetch("%%URL%%/my-plate/keys/command", {
  method: "POST",
  headers: {
    "Authorization": "YOUR_API_KEY",
    "Content-Type": "application/json"
  },
  body: JSON.stringify({
    command: "RENAME",
    args: ["oldkey", "newkey"]
  })
});
// result: "OK"
```

---

## COPY - Copy a Key

Copy a key to a new name (leaves original intact).

```javascript
await fetch("%%URL%%/my-plate/keys/command", {
  method: "POST",
  headers: {
    "Authorization": "YOUR_API_KEY",
    "Content-Type": "application/json"
  },
  body: JSON.stringify({
    command: "COPY",
    args: ["source", "destination"]
  })
});
// result: 1 (success)
```

---

## SCAN - Iterate Over Keys

Use SCAN to iterate through all keys without blocking.

### Basic Scan

```javascript
const response = await fetch("%%URL%%/my-plate/scan?cursor=0&count=100", {
  headers: { "Authorization": "YOUR_API_KEY" }
});
// result: { cursor: "10", keys: ["key1", "key2"], done: false }
```

### cURL

```bash
# Initial scan
curl -X GET "%%URL%%/my-plate/scan?cursor=0&count=100" \
  -H "Authorization: YOUR_API_KEY"

# Follow-up scan using returned cursor
curl -X GET "%%URL%%/my-plate/scan?cursor=10&count=100" \
  -H "Authorization: YOUR_API_KEY"
```

### Scan with Pattern Filter

Only return keys matching a pattern (glob syntax).

```bash
# Get all keys starting with "user:"
curl -X GET "%%URL%%/my-plate/scan?cursor=0&count=100&match=user:*" \
  -H "Authorization: YOUR_API_KEY"
```

### Scan by Key Type

Only return keys of a specific type.

```bash
# Only return string keys
curl -X GET "%%URL%%/my-plate/scan?cursor=0&count=100&type=string" \
  -H "Authorization: YOUR_API_KEY"

# Only return hash keys
curl -X GET "%%URL%%/my-plate/scan?cursor=0&count=100&type=hash" \
  -H "Authorization: YOUR_API_KEY"
```

Valid types: `string`, `list`, `set`, `zset`, `hash`, `stream`

### How Scan Works

1. Start with `cursor=0`
2. Get response with `nextCursor` and `keys`
3. If `done: false`, call again with next cursor
4. Repeat until `done: true`

---

## Collection Scanning

Scan members inside hashes, sets, or sorted sets.

### HSCAN - Scan Hash Fields

```javascript
await fetch("%%URL%%/my-plate/scan/hashes/user:1", {
  method: "POST",
  headers: {
    "Authorization": "YOUR_API_KEY",
    "Content-Type": "application/json"
  },
  body: JSON.stringify({
    cursor: 0,
    match: "f*",
    count: 10
  })
});
```

### SSCAN - Scan Set Members

```javascript
await fetch("%%URL%%/my-plate/scan/sets/myset", {
  method: "POST",
  headers: {
    "Authorization": "YOUR_API_KEY",
    "Content-Type": "application/json"
  },
  body: JSON.stringify({
    cursor: 0,
    count: 10
  })
});
```

### ZSCAN - Scan Sorted Set Members

```javascript
await fetch("%%URL%%/my-plate/scan/zsets/myzset", {
  method: "POST",
  headers: {
    "Authorization": "YOUR_API_KEY",
    "Content-Type": "application/json"
  },
  body: JSON.stringify({
    cursor: 0,
    count: 10
  })
});
```

---

## DELETE /{plateID}/keys/{pattern}

Delete multiple keys matching a pattern using glob syntax.

### Examples

```bash
# Delete all keys starting with "temp:"
curl -X DELETE "%%URL%%/my-plate/keys/temp:*" \
  -H "Authorization: YOUR_API_KEY"
# result: { deleted: 5, pattern: "temp:*" }

# Delete all keys ending with ":cache"
curl -X DELETE "%%URL%%/my-plate/keys/*:cache" \
  -H "Authorization: YOUR_API_KEY"
```

### Pattern Matching Guide

| Pattern | Matches | Example |
|---------|---------|---------|
| `*` | Any characters | `user:*` matches `user:1`, `user:abc` |
| `?` | Single character | `key?` matches `key1`, `keya` |
| `[]` | Character set | `key[12]` matches `key1`, `key2` |

---

## Command Reference

| Command | Description | Example |
|---------|-------------|---------|
| GET | Get value | `["key"]` |
| SET | Set value | `["key", "value", "EX", "3600"]` |
| MGET | Get multiple | `["key1", "key2"]` |
| MSET | Set multiple | `["key1", "val1", "key2", "val2"]` |
| DEL | Delete | `["key1", "key2"]` |
| UNLINK | Async delete | `["key1"]` |
| EXISTS | Check existence | `["key1", "key2"]` |
| TYPE | Get data type | `["key"]` |
| RENAME | Rename key | `["old", "new"]` |
| COPY | Copy key | `["src", "dst"]` |
| EXPIRE | Set TTL (sec) | `["key", "3600"]` |
| PEXPIRE | Set TTL (ms) | `["key", "3600000"]` |
| EXPIREAT | Expire at timestamp | `["key", "1735689600"]` |
| PEXPIREAT | Expire at ms timestamp | `["key", "1735689600000"]` |
| TTL | Get TTL (sec) | `["key"]` |
| PTTL | Get TTL (ms) | `["key"]` |
| PERSIST | Remove TTL | `["key"]` |
| GETEX | Get + set TTL | `["key", "EX", "3600"]` |
| GETDEL | Get + delete | `["key"]` |
| SCAN | Iterate keys | `["0", "COUNT", "100"]` |
