# Lists API

The Lists API provides operations for **list data types**. A list is an ordered collection of values, like an array. You can add items to the beginning or end, and access by position.

## Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/{plateID}/lists/command` | Execute list commands |
| POST | `/{plateID}/lists/{key}/command` | Execute key-specific list commands |

---

## What is a List?

A list is an ordered sequence of strings. Unlike sets, lists allow duplicates and maintain insertion order.

Think of it like:
- A queue (first in, first out)
- A stack (last in, first out)
- A todo list with order

---

## LPUSH - Add to Beginning

Add one or more elements to the **left (beginning)** of the list.

**Request:**
```json
{
  "command": "LPUSH",
  "args": ["mylist", "a", "b", "c"]
}
```

**Response:**
```json
{
  "result": 3
}
```
Returns the new length of the list.

### JavaScript

```javascript
const response = await fetch("%%URL%%/my-plate/lists/command", {
  method: "POST",
  headers: {
    "Authorization": "YOUR_API_KEY",
    "Content-Type": "application/json"
  },
  body: JSON.stringify({
    command: "LPUSH",
    args: ["mylist", "a", "b", "c"]
  })
});
// result: 3
```

### cURL

```bash
curl -X POST "%%URL%%/my-plate/lists/command" \
  -H "Authorization: YOUR_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{"command":"LPUSH","args":["mylist","a","b","c"]}'
```

### Visual Example

```
LPUSH mylist "a"  -> [a]
LPUSH mylist "b"  -> [b, a]
LPUSH mylist "c"  -> [c, b, a]
```

---

## RPUSH - Add to End

Add one or more elements to the **right (end)** of the list.

**Request:**
```json
{
  "command": "RPUSH",
  "args": ["mylist", "x", "y", "z"]
}
```

**Response:**
```json
{
  "result": 6
}
```

### cURL

```bash
curl -X POST "%%URL%%/my-plate/lists/command" \
  -H "Authorization: YOUR_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{"command":"RPUSH","args":["mylist","x","y","z"]}'
```

### Visual Example

```
RPUSH mylist "x" -> [c, b, a, x]
RPUSH mylist "y" -> [c, b, a, x, y]
RPUSH mylist "z" -> [c, b, a, x, y, z]
```

---

## LPOP - Remove from Beginning

Remove and return the **first element**.

**Request:**
```json
{
  "command": "LPOP",
  "args": ["mylist"]
}
```

**Response:**
```json
{
  "result": "c"
}
```

### JavaScript

```javascript
const response = await fetch("%%URL%%/my-plate/lists/command", {
  method: "POST",
  headers: {
    "Authorization": "YOUR_API_KEY",
    "Content-Type": "application/json"
  },
  body: JSON.stringify({
    command: "LPOP",
    args: ["mylist"]
  })
});
// result: "c"
```

### cURL

```bash
curl -X POST "%%URL%%/my-plate/lists/command" \
  -H "Authorization: YOUR_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{"command":"LPOP","args":["mylist"]}'
```

---

## RPOP - Remove from End

Remove and return the **last element**.

**Request:**
```json
{
  "command": "RPOP",
  "args": ["mylist"]
}
```

**Response:**
```json
{
  "result": "z"
}
```

### cURL

```bash
curl -X POST "%%URL%%/my-plate/lists/command" \
  -H "Authorization: YOUR_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{"command":"RPOP","args":["mylist"]}'
```

---

## LRANGE - Get Elements by Range

Get elements from start to end index.

- Index 0 = first element
- Negative indices count from the end (-1 = last)

**Request:**
```json
{
  "command": "LRANGE",
  "args": ["mylist", "0", "2"]
}
```

**Response:**
```json
{
  "result": ["b", "a", "x"]
}
```

### JavaScript

```javascript
const response = await fetch("%%URL%%/my-plate/lists/command", {
  method: "POST",
  headers: {
    "Authorization": "YOUR_API_KEY",
    "Content-Type": "application/json"
  },
  body: JSON.stringify({
    command: "LRANGE",
    args: ["mylist", 0, 2]
  })
});
// result: ["b", "a", "x"]
```

### cURL

```bash
# Get first 3 elements
curl -X POST "%%URL%%/my-plate/lists/command" \
  -H "Authorization: YOUR_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{"command":"LRANGE","args":["mylist","0","2"]}'

# Get all elements
curl -X POST "%%URL%%/my-plate/lists/command" \
  -H "Authorization: YOUR_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{"command":"LRANGE","args":["mylist","0","-1"]}'
```

---

## LLEN - Get List Length

Get the number of elements in the list.

**Request:**
```json
{
  "command": "LLEN",
  "args": ["mylist"]
}
```

**Response:**
```json
{
  "result": 4
}
```

### cURL

```bash
curl -X POST "%%URL%%/my-plate/lists/command" \
  -H "Authorization: YOUR_API_KEY" \
  -H "Content-Type": application/json" \
  -d '{"command":"LLEN","args":["mylist"]}'
```

---

## LINDEX - Get by Index

Get element at a specific position.

**Request:**
```json
{
  "command": "LINDEX",
  "args": ["mylist", "0"]
}
```

**Response:**
```json
{
  "result": "b"
}
```

### cURL

```bash
curl -X POST "%%URL%%/my-plate/lists/command" \
  -H "Authorization: YOUR_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{"command":"LINDEX","args":["mylist","0"]}'
```

---

## LPOS - Find Element Index

Find the position of an element.

**Request:**
```json
{
  "command": "LPOS",
  "args": ["mylist", "a"]
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
curl -X POST "%%URL%%/my-plate/lists/command" \
  -H "Authorization: YOUR_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{"command":"LPOS","args":["mylist","a"]}'
```

---

## LSET - Set by Index

Replace element at a specific position.

**Request:**
```json
{
  "command": "LSET",
  "args": ["mylist", "0", "first"]
}
```

**Response:**
```json
{
  "result": "OK"
}
```

### cURL

```bash
curl -X POST "%%URL%%/my-plate/lists/command" \
  -H "Authorization: YOUR_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{"command":"LSET","args":["mylist","0","first"]}'
```

---

## LINSERT - Insert Element

Insert a new element before or after a pivot element.

**Request:**
```json
{
  "command": "LINSERT",
  "args": ["mylist", "BEFORE", "a", "new"]
}
```

**Response:**
```json
{
  "result": 5
}
```
Returns the new list length.

### JavaScript

```javascript
await fetch("%%URL%%/my-plate/lists/command", {
  method: "POST",
  headers: {
    "Authorization": "YOUR_API_KEY",
    "Content-Type": "application/json"
  },
  body: JSON.stringify({
    command: "LINSERT",
    args: ["mylist", "BEFORE", "a", "new"]
  })
});
// result: 5

// Or after
await fetch("%%URL%%/my-plate/lists/command", {
  method: "POST",
  headers: {
    "Authorization": "YOUR_API_KEY",
    "Content-Type": "application/json"
  },
  body: JSON.stringify({
    command: "LINSERT",
    args: ["mylist", "AFTER", "a", "new"]
  })
});
```

### cURL

```bash
# Insert BEFORE pivot
curl -X POST "%%URL%%/my-plate/lists/command" \
  -H "Authorization: YOUR_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{"command":"LINSERT","args":["mylist","BEFORE","a","new"]}'

# Insert AFTER pivot
curl -X POST "%%URL%%/my-plate/lists/command" \
  -H "Authorization: YOUR_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{"command":"LINSERT","args":["mylist","AFTER","a","new"]}'
```

---

## LREM - Remove Elements

Remove occurrences of an element.

**Request:**
```json
{
  "command": "LREM",
  "args": ["mylist", "2", "a"]
}
```

**Response:**
```json
{
  "result": 2
}
```
First number = how many to remove (0 = all), second = the value.

### cURL

```bash
# Remove first 2 occurrences of "a"
curl -X POST "%%URL%%/my-plate/lists/command" \
  -H "Authorization: YOUR_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{"command":"LREM","args":["mylist","2","a"]}'

# Remove all occurrences of "a"
curl -X POST "%%URL%%/my-plate/lists/command" \
  -H "Authorization: YOUR_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{"command":"LREM","args":["mylist","0","a"]}'
```

---

## LTRIM - Trim List

Keep only elements in a range, remove everything else.

**Request:**
```json
{
  "command": "LTRIM",
  "args": ["mylist", "0", "2"]
}
```

**Response:**
```json
{
  "result": "OK"
}
```

### Use Case: Implement a queue with fixed size

```javascript
// Add item, keep only last 100
await fetch("%%URL%%/my-plate/lists/command", {
  method: "POST",
  headers: {
    "Authorization": "YOUR_API_KEY",
    "Content-Type": "application/json"
  },
  body: JSON.stringify({
    command: "LTRIM",
    args: ["mylist", -100, -1]
  })
});
```

### cURL

```bash
curl -X POST "%%URL%%/my-plate/lists/command" \
  -H "Authorization: YOUR_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{"command":"LTRIM","args":["mylist","0","2"]}'
```

---

## LMOVE - Move Element

Move an element from one list to another, or within the same list.

**Request:**
```json
{
  "command": "LMOVE",
  "args": ["list1", "list2", "RIGHT", "LEFT"]
}
```

**Response:**
```json
{
  "result": "element"
}
```

Directions: `LEFT` or `RIGHT` for source, then `LEFT` or `RIGHT` for destination.

### cURL

```bash
# Move last element from list1 to front of list2
curl -X POST "%%URL%%/my-plate/lists/command" \
  -H "Authorization: YOUR_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{"command":"LMOVE","args":["list1","list2","RIGHT","LEFT"]}'
```

---

## Command Reference

| Command | Description | Example |
|---------|-------------|---------|
| LPUSH | Push to head | `["key", "val1", "val2"]` |
| RPUSH | Push to tail | `["key", "val1", "val2"]` |
| LPOP | Pop from head | `["key"]` |
| RPOP | Pop from tail | `["key"]` |
| LLEN | Get length | `["key"]` |
| LRANGE | Get range | `["key", "0", "-1"]` |
| LINDEX | Get by index | `["key", "0"]` |
| LPOS | Get index of value | `["key", "value"]` |
| LSET | Set by index | `["key", "0", "value"]` |
| LINSERT | Insert | `["key", "BEFORE", "pivot", "value"]` |
| LREM | Remove | `["key", "count", "value"]` |
| LTRIM | Trim | `["key", "start", "stop"]` |
| LMOVE | Move | `["src", "dst", "RIGHT", "LEFT"]` |