# Bitmaps API

The Bitmaps API provides operations for **bitmap data types**. Bitmaps are incredibly efficient for storing boolean data (yes/no, 0/1) for large sets of items.

## Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/{plateID}/bitmaps/command` | Execute bitmap commands |
| POST | `/{plateID}/bitmaps/{key}/command` | Execute key-specific bitmap commands |

---

## What are Bitmaps?

A bitmap is an array of bits. Each bit represents whether something is on (1) or off (0).

Use cases:
- User activity tracking (visited/not visited)
- Daily active users
- Feature flags
- Boolean user properties

**Why use bitmaps?**
- Extremely memory efficient: 1 billion bits = ~125MB
- Fast operations at the bit level

---

## SETBIT - Set a Bit

Set a specific bit to 0 or 1.

**Request:**
```json
{
  "command": "SETBIT",
  "args": ["mybitmap", "500", "1"]
}
```

**Response:**
```json
{
  "result": 0
}
```
Returns the old value of the bit.

### JavaScript

```javascript
const response = await fetch("%%URL%%/my-plate/bitmaps/command", {
  method: "POST",
  headers: {
    "Authorization": "YOUR_API_KEY",
    "Content-Type": "application/json"
  },
  body: JSON.stringify({
    command: "SETBIT",
    args: ["mybitmap", 500, 1]
  })
});
// result: 0
```

### cURL

```bash
curl -X POST "%%URL%%/my-plate/bitmaps/command" \
  -H "Authorization: YOUR_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{"command":"SETBIT","args":["mybitmap","500","1"]}'
```

### Real-World Example: Track Daily Login

```javascript
// User 12345 logged in today (day 0)
await fetch("%%URL%%/my-plate/bitmaps/command", {
  method: "POST",
  headers: {
    "Authorization": "YOUR_API_KEY",
    "Content-Type": "application/json"
  },
  body: JSON.stringify({
    command: "SETBIT",
    args: ["login:2024-01", 12345, 1]
  })
});
```

---

## GETBIT - Get a Bit

Get the value of a specific bit.

**Request:**
```json
{
  "command": "GETBIT",
  "args": ["mybitmap", "500"]
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
curl -X POST "%%URL%%/my-plate/bitmaps/command" \
  -H "Authorization: YOUR_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{"command":"GETBIT","args":["mybitmap","500"]}'
```

---

## BITCOUNT - Count Set Bits

Count how many bits are set to 1 in a range.

**Request:**
```json
{
  "command": "BITCOUNT",
  "args": ["mybitmap", "0", "100"]
}
```

**Response:**
```json
{
  "result": 50
}
```

### JavaScript

```javascript
const response = await fetch("%%URL%%/my-plate/bitmaps/command", {
  method: "POST",
  headers: {
    "Authorization": "YOUR_API_KEY",
    "Content-Type": "application/json"
  },
  body: JSON.stringify({
    command: "BITCOUNT",
    args: ["mybitmap", 0, 100]
  })
});
// result: 50
```

### cURL

```bash
# Count bits from offset 0 to 100
curl -X POST "%%URL%%/my-plate/bitmaps/command" \
  -H "Authorization: YOUR_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{"command":"BITCOUNT","args":["mybitmap","0","100"]}'

# Count entire bitmap (use -1 for end)
curl -X POST "%%URL%%/my-plate/bitmaps/command" \
  -H "Authorization: YOUR_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{"command":"BITCOUNT","args":["mybitmap","0","-1"]}'
```

### Real-World: Daily Active Users

```javascript
// Count users who logged in today
const response = await fetch("%%URL%%/my-plate/bitmaps/command", {
  method: "POST",
  headers: {
    "Authorization": "YOUR_API_KEY",
    "Content-Type": "application/json"
  },
  body: JSON.stringify({
    command: "BITCOUNT",
    args: ["login:2024-01-01", 0, -1]
  })
});
console.log(`${response.result} users were active today`);
```

---

## BITPOS - Find First Bit

Find the position of the first bit set to 0 or 1.

**Request:**
```json
{
  "command": "BITPOS",
  "args": ["mybitmap", "0", "0", "100"]
}
```

**Response:**
```json
{
  "result": 0
}
```

### cURL

```bash
# Find first 0 bit in range
curl -X POST "%%URL%%/my-plate/bitmaps/command" \
  -H "Authorization: YOUR_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{"command":"BITPOS","args":["mybitmap","0","0","100"]}'
```

---

## BITOP - Bitwise Operations

Perform bitwise operations (AND, OR, XOR, NOT) across multiple bitmaps.

**Request:**
```json
{
  "command": "BITOP",
  "args": ["AND", "result", "bitmap1", "bitmap2"]
}
```

**Response:**
```json
{
  "result": 1000
}
```
Number of bits in result.

### JavaScript

```javascript
const response = await fetch("%%URL%%/my-plate/bitmaps/command", {
  method: "POST",
  headers: {
    "Authorization": "YOUR_API_KEY",
    "Content-Type": "application/json"
  },
  body: JSON.stringify({
    command: "BITOP",
    args: ["AND", "result", "bitmap1", "bitmap2"]
  })
});
```

### cURL

```bash
# AND operation
curl -X POST "%%URL%%/my-plate/bitmaps/command" \
  -H "Authorization: YOUR_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{"command":"BITOP","args":["AND","result","bitmap1","bitmap2"]}'

# OR operation
curl -X POST "%%URL%%/my-plate/bitmaps/command" \
  -H "Authorization: YOUR_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{"command":"BITOP","args":["OR","result","bitmap1","bitmap2"]}'

# NOT operation (inverts one bitmap)
curl -X POST "%%URL%%/my-plate/bitmaps/command" \
  -H "Authorization: YOUR_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{"command":"BITOP","args":["NOT","result","bitmap1"]}'
```

### Real-World: Users Active Both Days

```javascript
// Find users active on both day1 and day2
await fetch("%%URL%%/my-plate/bitmaps/command", {
  method: "POST",
  headers: {
    "Authorization": "YOUR_API_KEY",
    "Content-Type": "application/json"
  },
  body: JSON.stringify({
    command: "BITOP",
    args: ["AND", "both_days", "login:2024-01-01", "login:2024-01-02"]
  })
});
// Now count how many
```

---

## BITFIELD - Advanced Bit Operations

Read and write integer values stored at specific bit offsets.

```javascript
// Get an 8-bit signed integer at offset 0
await fetch("%%URL%%/my-plate/bitmaps/command", {
  method: "POST",
  headers: {
    "Authorization": "YOUR_API_KEY",
    "Content-Type": "application/json"
  },
  body: JSON.stringify({
    command: "BITFIELD",
    args: ["mybitmap", "GET", "i8", 0]
  })
});
```

```bash
curl -X POST "%%URL%%/my-plate/bitmaps/command" \
  -H "Authorization: YOUR_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{"command":"BITFIELD","args":["mybitmap","GET","i8","0"]}'
```

---

## Command Reference

| Command | Description | Example |
|---------|-------------|---------|
| SETBIT | Set bit | `["key", "offset", "value"]` |
| GETBIT | Get bit | `["key", "offset"]` |
| BITCOUNT | Count bits | `["key", "start", "end"]` |
| BITPOS | Find bit | `["key", "bit", "start", "end"]` |
| BITOP | Bitwise ops | `["AND\|OR\|XOR\|NOT", "dest", "key1", ...]` |
| BITFIELD | Bitfield ops | `["key", "GET\|SET\|INCR", "type", "offset"]` |