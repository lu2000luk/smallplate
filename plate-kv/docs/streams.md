# Streams API

The Streams API provides operations for **stream data types**. Streams are like append-only logs - perfect for event sourcing, message queues, and time-series data!

## Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/{plateID}/streams/command` | Execute stream commands |
| POST | `/{plateID}/streams/{key}/command` | Execute key-specific stream commands |

---

## What is a Stream?

A stream is an append-only data structure that stores entries in the order they were added.

Each entry has:
- **Entry ID** - unique timestamp-based ID (like `1715000000000-0`)
- **Fields** - key-value pairs (like a hash)

Use cases:
- Event logs
- Message queues
- Activity feeds
- Time-series data

---

## XADD - Add Entry

Add a new entry to the stream.

**Request:**
```json
{
  "command": "XADD",
  "args": ["mystream", "*", "field1", "value1", "field2", "value2"]
}
```

**Response:**
```json
{
  "result": "1715000000000-0"
}
```
The `*` tells Redis to auto-generate an ID. Returns the generated ID.

### JavaScript

```javascript
const response = await fetch("%%URL%%/my-plate/streams/command", {
  method: "POST",
  headers: {
    "Authorization": "YOUR_API_KEY",
    "Content-Type": "application/json"
  },
  body: JSON.stringify({
    command: "XADD",
    args: ["mystream", "*", "field1", "value1", "field2", "value2"]
  })
});
// result: "1715000000000-0"
```

### cURL

```bash
curl -X POST "%%URL%%/my-plate/streams/command" \
  -H "Authorization: YOUR_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{"command":"XADD","args":["mystream","*","field1","value1","field2","value2"]}'
```

### Entry ID Format

IDs are `timestamp-sequence`. You can also specify manually:
```javascript
args: ["mystream", "1715000000000-0", "field", "value"]
```

---

## XLEN - Get Length

Get the number of entries in the stream.

**Request:**
```json
{
  "command": "XLEN",
  "args": ["mystream"]
}
```

**Response:**
```json
{
  "result": 5
}
```

### cURL

```bash
curl -X POST "%%URL%%/my-plate/streams/command" \
  -H "Authorization: YOUR_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{"command":"XLEN","args":["mystream"]}'
```

---

## XRANGE - Get Entries (Forward)

Get entries in a range, from start to end.

- `-` = first entry
- `+` = last entry

**Request:**
```json
{
  "command": "XRANGE",
  "args": ["mystream", "-", "+", "COUNT", "10"]
}
```

**Response:**
```json
{
  "result": [
    ["1715000000000-0", ["field1", "value1"]],
    ["1715000000001-0", ["field1", "value2"]]
  ]
}
```

### JavaScript

```javascript
const response = await fetch("%%URL%%/my-plate/streams/command", {
  method: "POST",
  headers: {
    "Authorization": "YOUR_API_KEY",
    "Content-Type": "application/json"
  },
  body: JSON.stringify({
    command: "XRANGE",
    args: ["mystream", "-", "+", "COUNT", 10]
  })
});
```

### cURL

```bash
curl -X POST "%%URL%%/my-plate/streams/command" \
  -H "Authorization: YOUR_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{"command":"XRANGE","args":["mystream","-","+","COUNT","10"]}'
```

---

## XREVRANGE - Get Entries (Reverse)

Get entries from end to start (newest first).

**Request:**
```json
{
  "command": "XREVRANGE",
  "args": ["mystream", "+", "-", "COUNT", "10"]
}
```

**Response:**
```json
{
  "result": [
    ["1715000000001-0", ["field1", "value2"]],
    ["1715000000000-0", ["field1", "value1"]]
  ]
}
```

### cURL

```bash
curl -X POST "%%URL%%/my-plate/streams/command" \
  -H "Authorization: YOUR_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{"command":"XREVRANGE","args":["mystream","+","-","COUNT","10"]}'
```

---

## XREAD - Read Entries

Read entries from one or more streams, starting after a specific ID.

**Request:**
```json
{
  "command": "XREAD",
  "args": ["COUNT", "2", "STREAMS", "mystream", "0"]
}
```

**Response:**
```json
{
  "result": [["mystream", [["1715000000000-0", ["field1", "value1"]]]]]
}
```

### cURL

```bash
curl -X POST "%%URL%%/my-plate/streams/command" \
  -H "Authorization: YOUR_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{"command":"XREAD","args":["COUNT","2","STREAMS","mystream","0"]}'
```

### Key Difference: XRANGE vs XREAD

- **XRANGE**: Get entries you already know exist (by ID range)
- **XREAD**: Poll for NEW entries (after a certain ID)

---

## XTRIM - Trim Stream

Remove old entries, keeping only the newest ones.

**Request:**
```json
{
  "command": "XTRIM",
  "args": ["mystream", "MINID", "1715000000000-0"]
}
```

**Response:**
```json
{
  "result": 5
}
```
Number of entries removed.

### cURL

```bash
# Keep entries from ID onwards
curl -X POST "%%URL%%/my-plate/streams/command" \
  -H "Authorization: YOUR_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{"command":"XTRIM","args":["mystream","MINID","1715000000000-0"]}'

# Keep last 1000 entries
curl -X POST "%%URL%%/my-plate/streams/command" \
  -H "Authorization: YOUR_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{"command":"XTRIM","args":["mystream","MAXLEN","1000"]}'
```

---

## XDEL - Delete Entry

Delete a specific entry by ID.

**Request:**
```json
{
  "command": "XDEL",
  "args": ["mystream", "1715000000000-0"]
}
```

**Response:**
```json
{
  "result": 1
}
```

### cURL

```bash
curl -X POST "%%URL%%/my-plate/streams/command" \
  -H "Authorization: YOUR_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{"command":"XDEL","args":["mystream","1715000000000-0"]}'
```

---

## XINFO - Stream Info

Get information about a stream.

**Request:**
```json
{
  "command": "XINFO",
  "args": ["STREAM", "mystream"]
}
```

**Response:**
```json
{
  "result": {
    "length": 5,
    "firstEntry": ["0-0", ["field", "value"]],
    "lastEntry": ["5-0", ["field", "value"]]
  }
}
```

### cURL

```bash
curl -X POST "%%URL%%/my-plate/streams/command" \
  -H "Authorization: YOUR_API_KEY" \
  -H "Content-Type": application/json" \
  -d '{"command":"XINFO","args":["STREAM","mystream"]}'
```

---

## Consumer Groups (Advanced)

Consumer groups allow multiple consumers to process stream entries together, ensuring each entry is processed exactly once.

### Create a Consumer Group

```bash
curl -X POST "%%URL%%/my-plate/streams/command" \
  -H "Authorization: YOUR_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{"command":"XGROUP","args":["CREATE","mystream","mygroup","0","MKSTREAM"]}'
```

### XREADGROUP - Read from Group

```bash
curl -X POST "%%URL%%/my-plate/streams/command" \
  -H "Authorization: YOUR_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{"command":"XREADGROUP","args":["GROUP","mygroup","consumer1","COUNT","10","STREAMS","mystream",">"]}'
```

### XACK - Acknowledge Processing

```bash
curl -X POST "%%URL%%/my-plate/streams/command" \
  -H "Authorization: YOUR_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{"command":"XACK","args":["mystream","mygroup","1715000000000-0"]}'
```

### XPENDING - Check Pending

```bash
curl -X POST "%%URL%%/my-plate/streams/command" \
  -H "Authorization: YOUR_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{"command":"XPENDING","args":["mystream","mygroup"]}'
```

---

## Command Reference

| Command | Description | Example |
|---------|-------------|---------|
| XADD | Add entry | `["key", "*", "field", "value"]` |
| XLEN | Get length | `["key"]` |
| XRANGE | Get range | `["key", "start", "stop", "COUNT", "n"]` |
| XREVRANGE | Get rev range | `["key", "start", "stop", "COUNT", "n"]` |
| XREAD | Read | `["COUNT", "n", "STREAMS", "key", "id"]` |
| XTRIM | Trim | `["key", "MINID", "id"]` |
| XDEL | Delete | `["key", "id"]` |
| XINFO | Info | `["STREAM", "key"]` |
| XGROUP | Group ops | `["CREATE", "key", "group", "id"]` |
| XREADGROUP | Read group | `["GROUP", "g", "c", "STREAMS", "k", ">"]` |
| XACK | Ack | `["key", "group", "id"]` |
| XPENDING | Pending | `["key", "group"]` |