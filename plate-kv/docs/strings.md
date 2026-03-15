# Strings API

The Strings API provides operations for string data types - the simplest key-value storage where each key holds a single text value.

## Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/{plateID}/strings/command` | Execute string commands |
| POST | `/{plateID}/strings/{key}/command` | Execute key-specific string commands |

---

## Understanding Command Arguments

When working with the Strings API, you'll pass commands with arguments. Here's a quick reference:

### Common Argument Suffixes

These modifiers control how commands behave:

| Suffix | Meaning | Example |
|--------|---------|---------|
| `EX seconds` | Set expiry in seconds | `SET key value EX 3600` = expire in 1 hour |
| `PX milliseconds` | Set expiry in milliseconds | `SET key value PX 60000` = expire in 1 minute |
| `EXAT timestamp` | Expire at Unix timestamp (seconds) | `SET key value EXAT 1735689600` |
| `PXAT timestamp` | Expire at Unix timestamp (ms) | `SET key value PXAT 1735689600000` |
| `NX` | Only set if key doesn't exist | `SET key value NX` |
| `XX` | Only set if key already exists | `SET key value XX` |
| `KEEPTTL` | Keep existing TTL | `SET key value KEEPTTL` |

### What is TTL?

**TTL (Time To Live)** is how long a key exists before it automatically gets deleted. It's like an expiration date.

- `EX 3600` = 3600 seconds = 1 hour
- `EX 60` = 60 seconds = 1 minute
- `PX 60000` = 60000 milliseconds = 1 minute

---

## SET - Store a Value

The most common command - stores a value in a key.

### Basic Usage

**Request:**
```json
{
  "command": "SET",
  "args": ["mykey", "hello world"]
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
const response = await fetch("%%URL%%/my-plate/strings/command", {
  method: "POST",
  headers: {
    "Authorization": "YOUR_API_KEY",
    "Content-Type": "application/json"
  },
  body: JSON.stringify({
    command: "SET",
    args: ["mykey", "hello world"]
  })
});
```

### cURL

```bash
curl -X POST "%%URL%%/my-plate/strings/command" \
  -H "Authorization: YOUR_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{"command":"SET","args":["mykey","hello world"]}'
```

### With Expiration (EX)

```javascript
// Expire in 1 hour (3600 seconds)
await fetch("%%URL%%/my-plate/strings/command", {
  method: "POST",
  headers: {
    "Authorization": "YOUR_API_KEY",
    "Content-Type": "application/json"
  },
  body: JSON.stringify({
    command: "SET",
    args: ["mykey", "hello world", "EX", "3600"]
  })
});
```

```bash
# Expire in 1 hour
curl -X POST "%%URL%%/my-plate/strings/command" \
  -H "Authorization: YOUR_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{"command":"SET","args":["mykey","hello world","EX","3600"]}'
```

### With NX (Only If Not Exists)

```javascript
// Only set if key doesn't exist (good for locks)
await fetch("%%URL%%/my-plate/strings/command", {
  method: "POST",
  headers: {
    "Authorization": "YOUR_API_KEY",
    "Content-Type": "application/json"
  },
  body: JSON.stringify({
    command: "SET",
    args: ["mykey", "value", "NX"]
  })
});
```

---

## GET - Retrieve a Value

Get the value stored in a key.

### JavaScript

```javascript
const response = await fetch("%%URL%%/my-plate/strings/command", {
  method: "POST",
  headers: {
    "Authorization": "YOUR_API_KEY",
    "Content-Type": "application/json"
  },
  body: JSON.stringify({
    command: "GET",
    args: ["mykey"]
  })
});
const data = await response.json();
// data.result = "hello world"
```

### cURL

```bash
curl -X POST "%%URL%%/my-plate/strings/command" \
  -H "Authorization: YOUR_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{"command":"GET","args":["mykey"]}'
```

---

## MGET - Get Multiple Values

Retrieve multiple values in one request.

### JavaScript

```javascript
const response = await fetch("%%URL%%/my-plate/strings/command", {
  method: "POST",
  headers: {
    "Authorization": "YOUR_API_KEY",
    "Content-Type": "application/json"
  },
  body: JSON.stringify({
    command: "MGET",
    args: ["key1", "key2", "key3"]
  })
});
// result: ["value1", "value2", null]
```

### cURL

```bash
curl -X POST "%%URL%%/my-plate/strings/command" \
  -H "Authorization: YOUR_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{"command":"MGET","args":["key1","key2","key3"]}'
```

---

## MSET - Set Multiple Values

Set multiple key-value pairs at once.

### JavaScript

```javascript
const response = await fetch("%%URL%%/my-plate/strings/command", {
  method: "POST",
  headers: {
    "Authorization": "YOUR_API_KEY",
    "Content-Type": "application/json"
  },
  body: JSON.stringify({
    command: "MSET",
    args: ["key1", "value1", "key2", "value2"]
  })
});
// result: "OK"
```

### cURL

```bash
curl -X POST "%%URL%%/my-plate/strings/command" \
  -H "Authorization: YOUR_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{"command":"MSET","args":["key1","value1","key2","value2"]}'
```

---

## INCR/DECR - Numeric Operations

Use strings as counters! These commands increment or decrement numeric values.

### INCR - Increment by 1

```javascript
// Start with counter = 0, then INCR makes it 1
await fetch("%%URL%%/my-plate/strings/command", {
  method: "POST",
  headers: {
    "Authorization": "YOUR_API_KEY",
    "Content-Type": "application/json"
  },
  body: JSON.stringify({
    command: "INCR",
    args: ["counter"]
  })
});
// result: 1
```

### INCRBY - Increment by Custom Amount

```javascript
// Add 10 to counter
await fetch("%%URL%%/my-plate/strings/command", {
  method: "POST",
  headers: {
    "Authorization": "YOUR_API_KEY",
    "Content-Type": "application/json"
  },
  body: JSON.stringify({
    command: "INCRBY",
    args: ["counter", 10]
  })
});
// result: 11
```

### DECRBY - Decrement by Custom Amount

```javascript
// Subtract 5 from counter
await fetch("%%URL%%/my-plate/strings/command", {
  method: "POST",
  headers: {
    "Authorization": "YOUR_API_KEY",
    "Content-Type": "application/json"
  },
  body: JSON.stringify({
    command: "DECRBY",
    args: ["counter", 5]
  })
});
// result: 6
```

### INCRBYFLOAT - Decimal Numbers

```javascript
// Add 1.5 to a float value
await fetch("%%URL%%/my-plate/strings/command", {
  method: "POST",
  headers: {
    "Authorization": "YOUR_API_KEY",
    "Content-Type": "application/json"
  },
  body: JSON.stringify({
    command: "INCRBYFLOAT",
    args: ["price", "1.5"]
  })
});
// result: "7.5"
```

### cURL Examples

```bash
# Increment counter
curl -X POST "%%URL%%/my-plate/strings/command" \
  -H "Authorization: YOUR_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{"command":"INCR","args":["counter"]}'

# Increment by 10
curl -X POST "%%URL%%/my-plate/strings/command" \
  -H "Authorization: YOUR_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{"command":"INCRBY","args":["counter","10"]}'
```

---

## APPEND - Add to Existing Value

Append text to the end of an existing string.

```javascript
const response = await fetch("%%URL%%/my-plate/strings/command", {
  method: "POST",
  headers: {
    "Authorization": "YOUR_API_KEY",
    "Content-Type": "application/json"
  },
  body: JSON.stringify({
    command: "APPEND",
    args: ["mykey", " world"]
  })
});
// result: 11 (new string length)
```

```bash
curl -X POST "%%URL%%/my-plate/strings/command" \
  -H "Authorization: YOUR_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{"command":"APPEND","args":["mykey"," world"]}'
```

---

## STRLEN - Get String Length

Get the length (number of characters) of a string.

```javascript
const response = await fetch("%%URL%%/my-plate/strings/command", {
  method: "POST",
  headers: {
    "Authorization": "YOUR_API_KEY",
    "Content-Type": "application/json"
  },
  body: JSON.stringify({
    command: "STRLEN",
    args: ["mykey"]
  })
});
// result: 11
```

```bash
curl -X POST "%%URL%%/my-plate/strings/command" \
  -H "Authorization: YOUR_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{"command":"STRLEN","args":["mykey"]}'
```

---

## GETRANGE/SETRANGE - Substring Operations

### GETRANGE - Get Part of String

Get characters from position start to end (0-indexed, negative counts from end).

```javascript
const response = await fetch("%%URL%%/my-plate/strings/command", {
  method: "POST",
  headers: {
    "Authorization": "YOUR_API_KEY",
    "Content-Type": "application/json"
  },
  body: JSON.stringify({
    command: "GETRANGE",
    args: ["mykey", 0, 4]
  })
});
// result: "hello" (first 5 characters)
```

### SETRANGE - Replace Part of String

Replace characters starting at offset.

```javascript
const response = await fetch("%%URL%%/my-plate/strings/command", {
  method: "POST",
  headers: {
    "Authorization": "YOUR_API_KEY",
    "Content-Type": "application/json"
  },
  body: JSON.stringify({
    command: "SETRANGE",
    args: ["mykey", 6, "world"]
  })
});
// result: 11
```

---

## GETEX - Get and Optionally Set Expiration

Get the value AND set expiration in one command.

```javascript
const response = await fetch("%%URL%%/my-plate/strings/command", {
  method: "POST",
  headers: {
    "Authorization": "YOUR_API_KEY",
    "Content-Type": "application/json"
  },
  body: JSON.stringify({
    command: "GETEX",
    args: ["mykey", "EX", "3600"]
  })
});
// result: "value" (and sets expiry to 1 hour)
```

```bash
# Get value and set 1 hour expiry
curl -X POST "%%URL%%/my-plate/strings/command" \
  -H "Authorization: YOUR_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{"command":"GETEX","args":["mykey","EX","3600"]}'
```

---

## GETDEL - Get and Delete

Get the value AND delete the key in one atomic operation.

```javascript
const response = await fetch("%%URL%%/my-plate/strings/command", {
  method: "POST",
  headers: {
    "Authorization": "YOUR_API_KEY",
    "Content-Type": "application/json"
  },
  body: JSON.stringify({
    command: "GETDEL",
    args: ["mykey"]
  })
});
// result: "value" (key is now deleted)
```

---

## Command Reference

| Command | Description | Example |
|---------|-------------|---------|
| SET | Set value | `["key", "value", "EX", "3600"]` |
| GET | Get value | `["key"]` |
| MGET | Get multiple | `["key1", "key2"]` |
| MSET | Set multiple | `["key1", "val1", "key2", "val2"]` |
| INCR | Increment by 1 | `["counter"]` |
| DECR | Decrement by 1 | `["counter"]` |
| INCRBY | Increment | `["counter", "10"]` |
| DECRBY | Decrement | `["counter", "10"]` |
| INCRBYFLOAT | Increment float | `["price", "1.5"]` |
| APPEND | Append | `["key", "suffix"]` |
| STRLEN | Length | `["key"]` |
| SETRANGE | Replace | `["key", "6", "world"]` |
| GETRANGE | Get substring | `["key", "0", "4"]` |
| GETEX | Get + expire | `["key", "EX", "3600"]` |
| GETDEL | Get + delete | `["key"]` |