# Sets API

The Sets API provides operations for **set data types**. A set is an unordered collection of unique strings - no duplicates allowed!

## Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/{plateID}/sets/command` | Execute set commands |
| POST | `/{plateID}/sets/{key}/command` | Execute key-specific set commands |

---

## What is a Set?

A set is like a mathematical set - a collection of unique items with no particular order.

Use cases:
- Tracking unique visitors
- Storing tags
- Managing friendships
- Any scenario where you need unique values

---

## SADD - Add Members

Add one or more members to a set.

**Request:**
```json
{
  "command": "SADD",
  "args": ["myset", "a", "b", "c"]
}
```

**Response:**
```json
{
  "result": 3
}
```
Number of new members added (duplicates are ignored).

### JavaScript

```javascript
const response = await fetch("%%URL%%/my-plate/sets/command", {
  method: "POST",
  headers: {
    "Authorization": "YOUR_API_KEY",
    "Content-Type": "application/json"
  },
  body: JSON.stringify({
    command: "SADD",
    args: ["myset", "a", "b", "c"]
  })
});
// result: 3
```

### cURL

```bash
curl -X POST "%%URL%%/my-plate/sets/command" \
  -H "Authorization: YOUR_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{"command":"SADD","args":["myset","a","b","c"]}'
```

### Adding Duplicate Values

```javascript
// Try to add duplicates
await fetch("%%URL%%/my-plate/sets/command", {
  method: "POST",
  headers: {
    "Authorization": "YOUR_API_KEY",
    "Content-Type": "application/json"
  },
  body: JSON.stringify({
    command: "SADD",
    args: ["myset", "a", "b", "a"]
  })
});
// result: 1 (only "a" is new!)
```

---

## SREM - Remove Members

Remove members from a set.

**Request:**
```json
{
  "command": "SREM",
  "args": ["myset", "b"]
}
```

**Response:**
```json
{
  "result": 1
}
```
Number of members removed.

### cURL

```bash
curl -X POST "%%URL%%/my-plate/sets/command" \
  -H "Authorization: YOUR_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{"command":"SREM","args":["myset","b"]}'
```

---

## SMEMBERS - Get All Members

Get all members of a set.

**Request:**
```json
{
  "command": "SMEMBERS",
  "args": ["myset"]
}
```

**Response:**
```json
{
  "result": ["a", "c"]
}
```

### JavaScript

```javascript
const response = await fetch("%%URL%%/my-plate/sets/command", {
  method: "POST",
  headers: {
    "Authorization": "YOUR_API_KEY",
    "Content-Type": "application/json"
  },
  body: JSON.stringify({
    command: "SMEMBERS",
    args: ["myset"]
  })
});
// result: ["a", "c"]
```

### cURL

```bash
curl -X POST "%%URL%%/my-plate/sets/command" \
  -H "Authorization: YOUR_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{"command":"SMEMBERS","args":["myset"]}'
```

---

## SISMEMBER - Check Membership

Check if a value is a member of the set.

**Request:**
```json
{
  "command": "SISMEMBER",
  "args": ["myset", "a"]
}
```

**Response:**
```json
{
  "result": 1
}
```
1 = yes, 0 = no.

### cURL

```bash
curl -X POST "%%URL%%/my-plate/sets/command" \
  -H "Authorization: YOUR_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{"command":"SISMEMBER","args":["myset","a"]}'
```

---

## SMISMEMBER - Check Multiple Members

Check membership for multiple values at once.

**Request:**
```json
{
  "command": "SMISMEMBER",
  "args": ["myset", "a", "b", "c"]
}
```

**Response:**
```json
{
  "result": [1, 0, 1]
}
```
Array of 1s and 0s for each check.

### cURL

```bash
curl -X POST "%%URL%%/my-plate/sets/command" \
  -H "Authorization: YOUR_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{"command":"SMISMEMBER","args":["myset","a","b","c"]}'
```

---

## SCARD - Get Size

Get the number of members in a set.

**Request:**
```json
{
  "command": "SCARD",
  "args": ["myset"]
}
```

**Response:**
```json
{
  "result": 2
}
```

### cURL

```bash
curl -X POST "%%URL%%/my-plate/sets/command" \
  -H "Authorization: YOUR_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{"command":"SCARD","args":["myset"]}'
```

---

## SRANDMEMBER - Get Random Member

Get one or more random members without removing them.

**Request:**
```json
{
  "command": "SRANDMEMBER",
  "args": ["myset", "2"]
}
```

**Response:**
```json
{
  "result": ["a", "c"]
}
```

### cURL

```bash
curl -X POST "%%URL%%/my-plate/sets/command" \
  -H "Authorization: YOUR_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{"command":"SRANDMEMBER","args":["myset","2"]}'
```

---

## SPOP - Pop Random Member

Remove and return random member(s).

**Request:**
```json
{
  "command": "SPOP",
  "args": ["myset"]
}
```

**Response:**
```json
{
  "result": "a"
}
```

### cURL

```bash
curl -X POST "%%URL%%/my-plate/sets/command" \
  -H "Authorization: YOUR_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{"command":"SPOP","args":["myset"]}'
```

---

## Set Operations

### SUNION - Union (Combine)

Get all members from multiple sets (no duplicates).

**Request:**
```json
{
  "command": "SUNION",
  "args": ["set1", "set2"]
}
```

**Response:**
```json
{
  "result": ["a", "b", "c", "d"]
}
```

### JavaScript

```javascript
const response = await fetch("%%URL%%/my-plate/sets/command", {
  method: "POST",
  headers: {
    "Authorization": "YOUR_API_KEY",
    "Content-Type": "application/json"
  },
  body: JSON.stringify({
    command: "SUNION",
    args: ["set1", "set2"]
  })
});
// result: ["a", "b", "c", "d"]
```

### cURL

```bash
curl -X POST "%%URL%%/my-plate/sets/command" \
  -H "Authorization: YOUR_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{"command":"SUNION","args":["set1","set2"]}'
```

---

### SINTER - Intersection (Common)

Get members that exist in ALL sets.

**Request:**
```json
{
  "command": "SINTER",
  "args": ["set1", "set2"]
}
```

**Response:**
```json
{
  "result": ["b"]
}
```

### cURL

```bash
curl -X POST "%%URL%%/my-plate/sets/command" \
  -H "Authorization: YOUR_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{"command":"SINTER","args":["set1","set2"]}'
```

---

### SDIFF - Difference (Only in First)

Get members only in the first set, not in others.

**Request:**
```json
{
  "command": "SDIFF",
  "args": ["set1", "set2"]
}
```

**Response:**
```json
{
  "result": ["a"]
}
```

### cURL

```bash
curl -X POST "%%URL%%/my-plate/sets/command" \
  -H "Authorization: YOUR_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{"command":"SDIFF","args":["set1","set2"]}'
```

---

## Store Operations

Save set operation results to a new key.

### SUNIONSTORE

```bash
curl -X POST "%%URL%%/my-plate/sets/command" \
  -H "Authorization: YOUR_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{"command":"SUNIONSTORE","args":["dest","set1","set2"]}'
```

### SINTERSTORE

```bash
curl -X POST "%%URL%%/my-plate/sets/command" \
  -H "Authorization: YOUR_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{"command":"SINTERSTORE","args":["dest","set1","set2"]}'
```

### SDIFFSTORE

```bash
curl -X POST "%%URL%%/my-plate/sets/command" \
  -H "Authorization: YOUR_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{"command":"SDIFFSTORE","args":["dest","set1","set2"]}'
```

---

## Command Reference

| Command | Description | Example |
|---------|-------------|---------|
| SADD | Add members | `["key", "val1", "val2"]` |
| SREM | Remove members | `["key", "val1"]` |
| SMEMBERS | Get all | `["key"]` |
| SISMEMBER | Check member | `["key", "value"]` |
| SMISMEMBER | Check multiple | `["key", "v1", "v2"]` |
| SCARD | Get count | `["key"]` |
| SRANDMEMBER | Random member | `["key", "count"]` |
| SPOP | Pop random | `["key", "count"]` |
| SUNION | Union | `["key1", "key2"]` |
| SINTER | Intersection | `["key1", "key2"]` |
| SDIFF | Difference | `["key1", "key2"]` |
| SUNIONSTORE | Store union | `["dest", "key1", "key2"]` |
| SINTERSTORE | Store inter | `["dest", "key1", "key2"]` |
| SDIFFSTORE | Store diff | `["dest", "key1", "key2"]` |