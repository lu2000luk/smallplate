# Pipeline & Transaction API

The Pipeline and Transaction API provides **batch execution** of multiple commands. This dramatically improves performance when you need to run many commands!

## Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/{plateID}/pipeline` | Execute multiple commands in pipeline |
| POST | `/{plateID}/transaction` | Execute multiple commands in transaction |

---

## Why Use Pipeline/Transaction?

### The Problem

Each HTTP request has network overhead. If you need to run 100 commands:
- Naive: 100 HTTP requests = lots of waiting
- Pipeline: 1 HTTP request with 100 commands = fast!

### The Solution

- **Pipeline**: Send multiple commands in one request, get all results at once
- **Transaction**: Pipeline + atomic execution (all or nothing)

---

## Pipeline - Batch Commands

Execute multiple commands in sequence. If one fails, others still run!

### POST /{plateID}/pipeline

**Request:**
```json
[
  ["SET", "key1", "value1"],
  ["GET", "key1"],
  ["INCR", "counter"],
  ["HGET", "user:1", "name"]
]
```

**Response:**
```json
[
  "OK",
  "value1",
  1,
  "John"
]
```

### JavaScript

```javascript
const response = await fetch("%%URL%%/my-plate/pipeline", {
  method: "POST",
  headers: {
    "Authorization": "YOUR_API_KEY",
    "Content-Type": "application/json"
  },
  body: JSON.stringify([
    ["SET", "key1", "value1"],
    ["GET", "key1"],
    ["INCR", "counter"],
    ["HGET", "user:1", "name"]
  ])
});
// result: ["OK", "value1", 1, "John"]
```

### cURL

```bash
curl -X POST "%%URL%%/my-plate/pipeline" \
  -H "Authorization: YOUR_API_KEY" \
  -H "Content-Type: application/json" \
  -d '[["SET","key1","value1"],["GET","key1"],["INCR","counter"],["HGET","user:1","name"]]'
```

### Use Cases

- **Bulk loading**: Import lots of data
- **Batch reads**: Get many values at once
- **Multiple writes**: Update many keys at once

---

## Transaction - Atomic Execution

Execute multiple commands atomically. Either ALL succeed, or NONE are applied!

### POST /{plateID}/transaction

**Request:**
```json
[
  ["INCR", "balance"],
  ["DECR", "debt"],
  ["SET", "lastTransaction", "2024-01-01"]
]
```

**Response:**
```json
[
  100,
  50,
  "OK"
]
```

### JavaScript

```javascript
const response = await fetch("%%URL%%/my-plate/transaction", {
  method: "POST",
  headers: {
    "Authorization": "YOUR_API_KEY",
    "Content-Type": "application/json"
  },
  body: JSON.stringify([
    ["INCR", "balance"],
    ["DECR", "debt"],
    ["SET", "lastTransaction", "2024-01-01"]
  ])
});
// result: [100, 50, "OK"]
```

### cURL

```bash
curl -X POST "%%URL%%/my-plate/transaction" \
  -H "Authorization: YOUR_API_KEY" \
  -H "Content-Type: application/json" \
  -d '[["INCR","balance"],["DECR","debt"],["SET","lastTransaction","2024-01-01"]]'
```

---

## Pipeline vs Transaction

| Feature | Pipeline | Transaction |
|---------|----------|-------------|
| Atomic | No | Yes |
| Rollback on failure | No | Yes |
| Performance gain | High | Medium |
| Use case | Read-heavy bulk ops | Write-heavy atomic ops |

---

## When to Use Each

### Use Pipeline When:
- Doing bulk operations
- Performance is critical
- Individual failures don't affect each other

### Use Transaction When:
- Need atomic operations
- Dependent updates (if one fails, all should fail)
- Financial operations, inventory, etc.

---

## Examples

### Pipeline: Bulk Set and Get

```javascript
await fetch("%%URL%%/my-plate/pipeline", {
  method: "POST",
  headers: {
    "Authorization": "YOUR_API_KEY",
    "Content-Type": "application/json"
  },
  body: JSON.stringify([
    ["MSET", "key1", "val1", "key2", "val2", "key3", "val3"],
    ["MGET", "key1", "key2", "key3"]
  ])
});
// result: ["OK", ["val1", "val2", "val3"]]
```

### cURL

```bash
curl -X POST "%%URL%%/my-plate/pipeline" \
  -H "Authorization: YOUR_API_KEY" \
  -H "Content-Type: application/json" \
  -d '[["MSET","key1","val1","key2","val2","key3","val3"],["MGET","key1","key2","key3"]]'
```

### Pipeline: Multiple Data Types

```javascript
await fetch("%%URL%%/my-plate/pipeline", {
  method: "POST",
  headers: {
    "Authorization": "YOUR_API_KEY",
    "Content-Type": "application/json"
  },
  body: JSON.stringify([
    ["SET", "string:key", "value"],
    ["HSET", "hash:key", "field", "value"],
    ["SADD", "set:key", "member"],
    ["ZADD", "zset:key", "100", "member"],
    ["LPUSH", "list:key", "item"]
  ])
});
// result: [1, 1, 1, 1, 1]
```

### Transaction: Account Transfer

```javascript
await fetch("%%URL%%/my-plate/transaction", {
  method: "POST",
  headers: {
    "Authorization": "YOUR_API_KEY",
    "Content-Type": "application/json"
  },
  body: JSON.stringify([
    ["DECRBY", "account:a", "100"],
    ["INCRBY", "account:b", "100"],
    ["SET", "transfer:completed", "true"]
  ])
});
```

### cURL

```bash
curl -X POST "%%URL%%/my-plate/transaction" \
  -H "Authorization: YOUR_API_KEY" \
  -H "Content-Type: application/json" \
  -d '[["DECRBY","account:a","100"],["INCRBY","account:b","100"],["SET","transfer:completed","true"]]'
```

---

## Error Handling

### Pipeline Errors

If a command fails, the response includes the error, but other commands still run:

```javascript
// Request
[
  ["SET", "key1", "value1"],
  ["GET", "nonexistent"],
  ["INCR", "notanumber"]
]

// Response
[
  "OK",
  null,
  {"error": "ERR value is not an integer or out of range"}
]
```

### Transaction Errors

If any command in a transaction fails, the entire transaction fails:

```javascript
// Request - if INCR fails, entire transaction is rolled back
[
  ["SET", "key1", "value1"],
  ["INCR", "notanumber"]
]

// Response
{"error": "EXECABORTED", "message": "..."}
```

---

## Best Practices

1. **Batch size**: Don't send thousands of commands at once
2. **Pipeline for reads**: Great for reducing latency on multiple reads
3. **Transaction for writes**: Use when atomicity matters
4. **Check results**: Always check response for errors