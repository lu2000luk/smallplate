# Hashes API

The Hashes API provides operations for **hash data types**. Think of a hash like a JavaScript object or Python dictionary - it's a collection of field-value pairs inside a single key.

## Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/{plateID}/hashes/command` | Execute hash commands |
| POST | `/{plateID}/hashes/{key}/command` | Execute key-specific hash commands |

---

## What is a Hash?

A hash is perfect for storing objects or records. For example, a user record:

```json
{
  "name": "John",
  "age": "30",
  "email": "john@example.com"
}
```

This would be stored as a single key (like `user:1`) with multiple fields inside it.

---

## HSET - Set Fields

Add or update field-value pairs in a hash.

**Request:**
```json
{
  "command": "HSET",
  "args": ["user:1", "name", "John", "age", "30"]
}
```

**Response:**
```json
{
  "result": 2
}
```
The result is the number of fields that were set.

### JavaScript

```javascript
const response = await fetch("%%URL%%/my-plate/hashes/command", {
  method: "POST",
  headers: {
    "Authorization": "YOUR_API_KEY",
    "Content-Type": "application/json"
  },
  body: JSON.stringify({
    command: "HSET",
    args: ["user:1", "name", "John", "age", "30"]
  })
});
// result: 2
```

### cURL

```bash
curl -X POST "%%URL%%/my-plate/hashes/command" \
  -H "Authorization: YOUR_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{"command":"HSET","args":["user:1","name","John","age","30"]}'
```

---

## HGET - Get One Field

Get a single field's value.

**Request:**
```json
{
  "command": "HGET",
  "args": ["user:1", "name"]
}
```

**Response:**
```json
{
  "result": "John"
}
```

### JavaScript

```javascript
const response = await fetch("%%URL%%/my-plate/hashes/command", {
  method: "POST",
  headers: {
    "Authorization": "YOUR_API_KEY",
    "Content-Type": "application/json"
  },
  body: JSON.stringify({
    command: "HGET",
    args: ["user:1", "name"]
  })
});
```

### cURL

```bash
curl -X POST "%%URL%%/my-plate/hashes/command" \
  -H "Authorization: YOUR_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{"command":"HGET","args":["user:1","name"]}'
```

---

## HMGET - Get Multiple Fields

Get values for multiple fields at once.

**Request:**
```json
{
  "command": "HMGET",
  "args": ["user:1", "name", "age", "email"]
}
```

**Response:**
```json
{
  "result": ["John", "30", "john@example.com"]
}
```

### JavaScript

```javascript
const response = await fetch("%%URL%%/my-plate/hashes/command", {
  method: "POST",
  headers: {
    "Authorization": "YOUR_API_KEY",
    "Content-Type": "application/json"
  },
  body: JSON.stringify({
    command: "HMGET",
    args: ["user:1", "name", "age", "email"]
  })
});
// result: ["John", "30", "john@example.com"]
```

### cURL

```bash
curl -X POST "%%URL%%/my-plate/hashes/command" \
  -H "Authorization: YOUR_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{"command":"HMGET","args":["user:1","name","age","email"]}'
```

---

## HGETALL - Get All Fields and Values

Get every field and value in the hash.

**Request:**
```json
{
  "command": "HGETALL",
  "args": ["user:1"]
}
```

**Response:**
```json
{
  "result": ["name", "John", "age", "30", "email", "john@example.com"]
}
```
Returns as alternating field, value, field, value...

### JavaScript

```javascript
const response = await fetch("%%URL%%/my-plate/hashes/command", {
  method: "POST",
  headers: {
    "Authorization": "YOUR_API_KEY",
    "Content-Type": "application/json"
  },
  body: JSON.stringify({
    command: "HGETALL",
    args: ["user:1"]
  })
});
// result: ["name", "John", "age", "30", "email", "john@example.com"]
```

### cURL

```bash
curl -X POST "%%URL%%/my-plate/hashes/command" \
  -H "Authorization: YOUR_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{"command":"HGETALL","args":["user:1"]}'
```

---

## HKEYS - Get All Field Names

Get only the field names (not values).

**Request:**
```json
{
  "command": "HKEYS",
  "args": ["user:1"]
}
```

**Response:**
```json
{
  "result": ["name", "age", "email"]
}
```

### cURL

```bash
curl -X POST "%%URL%%/my-plate/hashes/command" \
  -H "Authorization: YOUR_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{"command":"HKEYS","args":["user:1"]}'
```

---

## HVALS - Get All Values

Get only the values (not field names).

**Request:**
```json
{
  "command": "HVALS",
  "args": ["user:1"]
}
```

**Response:**
```json
{
  "result": ["John", "30", "john@example.com"]
}
```

### cURL

```bash
curl -X POST "%%URL%%/my-plate/hashes/command" \
  -H "Authorization: YOUR_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{"command":"HVALS","args":["user:1"]}'
```

---

## HDEL - Delete Fields

Remove specific fields from a hash.

**Request:**
```json
{
  "command": "HDEL",
  "args": ["user:1", "age"]
}
```

**Response:**
```json
{
  "result": 1
}
```
Number of fields deleted.

### JavaScript

```javascript
await fetch("%%URL%%/my-plate/hashes/command", {
  method: "POST",
  headers: {
    "Authorization": "YOUR_API_KEY",
    "Content-Type": "application/json"
  },
  body: JSON.stringify({
    command: "HDEL",
    args: ["user:1", "age"]
  })
});
// result: 1
```

### cURL

```bash
curl -X POST "%%URL%%/my-plate/hashes/command" \
  -H "Authorization: YOUR_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{"command":"HDEL","args":["user:1","age"]}'
```

---

## HEXISTS - Check if Field Exists

Check if a specific field exists in the hash.

**Request:**
```json
{
  "command": "HEXISTS",
  "args": ["user:1", "name"]
}
```

**Response:**
```json
{
  "result": 1
}
```
1 = exists, 0 = doesn't exist.

### cURL

```bash
curl -X POST "%%URL%%/my-plate/hashes/command" \
  -H "Authorization: YOUR_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{"command":"HEXISTS","args":["user:1","name"]}'
```

---

## HLEN - Get Field Count

Get the number of fields in a hash.

**Request:**
```json
{
  "command": "HLEN",
  "args": ["user:1"]
}
```

**Response:**
```json
{
  "result": 3
}
```

### cURL

```bash
curl -X POST "%%URL%%/my-plate/hashes/command" \
  -H "Authorization: YOUR_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{"command":"HLEN","args":["user:1"]}'
```

---

## HINCRBY - Increment a Number

Increment an integer field value. Great for counters!

**Request:**
```json
{
  "command": "HINCRBY",
  "args": ["user:1", "login_count", 1]
}
```

**Response:**
```json
{
  "result": 1
}
```

### JavaScript

```javascript
// Increment login count
await fetch("%%URL%%/my-plate/hashes/command", {
  method: "POST",
  headers: {
    "Authorization": "YOUR_API_KEY",
    "Content-Type": "application/json"
  },
  body: JSON.stringify({
    command: "HINCRBY",
    args: ["user:1", "login_count", 1]
  })
});
// result: 1

// Increment by 10
await fetch("%%URL%%/my-plate/hashes/command", {
  method: "POST",
  headers: {
    "Authorization": "YOUR_API_KEY",
    "Content-Type": "application/json"
  },
  body: JSON.stringify({
    command: "HINCRBY",
    args: ["user:1", "points", 10]
  })
});
```

### cURL

```bash
# Increment by 1
curl -X POST "%%URL%%/my-plate/hashes/command" \
  -H "Authorization: YOUR_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{"command":"HINCRBY","args":["user:1","login_count","1"]}'

# Increment by 10
curl -X POST "%%URL%%/my-plate/hashes/command" \
  -H "Authorization: YOUR_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{"command":"HINCRBY","args":["user:1","points","10"]}'
```

---

## HINCRBYFLOAT - Increment a Decimal

Increment a float (decimal) field value.

**Request:**
```json
{
  "command": "HINCRBYFLOAT",
  "args": ["product:1", "price", "0.50"]
}
```

**Response:**
```json
{
  "result": "10.50"
}
```

### cURL

```bash
curl -X POST "%%URL%%/my-plate/hashes/command" \
  -H "Authorization: YOUR_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{"command":"HINCRBYFLOAT","args":["product:1","price","0.50"]}'
```

---

## HSETNX - Set if Not Exists

Only set a field if it doesn't already exist. Good for initialization.

**Request:**
```json
{
  "command": "HSETNX",
  "args": ["user:1", "nickname", "Johnny"]
}
```

**Response:**
```json
{
  "result": 1
}
```
1 = set, 0 = already existed.

### cURL

```bash
curl -X POST "%%URL%%/my-plate/hashes/command" \
  -H "Authorization: YOUR_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{"command":"HSETNX","args":["user:1","nickname","Johnny"]}'
```

---

## HRANDRANDOM - Get Random Fields

Get random field(s) from the hash.

**Request:**
```json
{
  "command": "HRANDFIELD",
  "args": ["user:1", "2"]
}
```

**Response:**
```json
{
  "result": ["name", "age"]
}
```

### cURL

```bash
curl -X POST "%%URL%%/my-plate/hashes/command" \
  -H "Authorization: YOUR_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{"command":"HRANDFIELD","args":["user:1","2"]}'
```

---

## Command Reference

| Command | Description | Example |
|---------|-------------|---------|
| HSET | Set field(s) | `["key", "field", "value", ...]` |
| HGET | Get field | `["key", "field"]` |
| HMGET | Get fields | `["key", "field1", "field2"]` |
| HGETALL | Get all | `["key"]` |
| HDEL | Delete fields | `["key", "field1", "field2"]` |
| HEXISTS | Check existence | `["key", "field"]` |
| HINCRBY | Increment int | `["key", "field", "10"]` |
| HINCRBYFLOAT | Increment float | `["key", "field", "1.5"]` |
| HKEYS | Get field names | `["key"]` |
| HVALS | Get values | `["key"]` |
| HLEN | Get count | `["key"]` |
| HSETNX | Set if not exists | `["key", "field", "value"]` |
| HRANDRANDOM | Random fields | `["key", "count"]` |