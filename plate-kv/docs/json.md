# JSON API

The JSON API provides specialized endpoints for storing and retrieving **JSON data** directly. This is simpler than storing JSON as a string!

## Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/{plateID}/json/{key}` | Get JSON value |
| POST | `/{plateID}/json/{key}` | Set JSON value |

---

## Why Use JSON Endpoints?

When storing JSON data, you could use regular strings:
- You need to manually serialize/deserialize
- No type checking

JSON endpoints give you:
- Automatic JSON parsing
- Type-safe storage
- Cleaner API

---

## GET /{plateID}/json/{key}

Get a JSON value by key.

### JavaScript

```javascript
const response = await fetch("%%URL%%/my-plate/json/user:1", {
  headers: {
    "Authorization": "YOUR_API_KEY"
  }
});
const data = await response.json();
// { name: "John", age: 30, active: true }
```

### cURL

```bash
curl -X GET "%%URL%%/my-plate/json/user:1" \
  -H "Authorization: YOUR_API_KEY"
```

---

## POST /{plateID}/json/{key}

Store a JSON value with optional expiration (TTL).

**Request:**
```json
{
  "value": {
    "name": "John",
    "age": 30,
    "active": true
  },
  "ttl_ms": 3600000
}
```

**Response:**
```json
{
  "stored": true
}
```

### JavaScript

```javascript
const response = await fetch("%%URL%%/my-plate/json/user:1", {
  method: "POST",
  headers: {
    "Authorization": "YOUR_API_KEY",
    "Content-Type": "application/json"
  },
  body: JSON.stringify({
    value: {
      name: "John",
      age: 30,
      active: true
    },
    ttl_ms: 3600000 // 1 hour in milliseconds
  })
});
// result: { stored: true }
```

### cURL

```bash
curl -X POST "%%URL%%/my-plate/json/user:1" \
  -H "Authorization: YOUR_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{"value":{"name":"John","age":30,"active":true},"ttl_ms":3600000}'
```

---

## Storing Without TTL

Omit `ttl_ms` for keys that never expire:

### JavaScript

```javascript
await fetch("%%URL%%/my-plate/json/config", {
  method: "POST",
  headers: {
    "Authorization": "YOUR_API_KEY",
    "Content-Type": "application/json"
  },
  body: JSON.stringify({
    value: {
      theme: "dark",
      notifications: true,
      version: "1.0.0"
    }
  })
});
```

### cURL

```bash
curl -X POST "%%URL%%/my-plate/json/config" \
  -H "Authorization: YOUR_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{"value":{"theme":"dark","notifications":true,"version":"1.0.0"}}'
```

---

## Storing Different JSON Types

### Object

```javascript
await fetch("%%URL%%/my-plate/json/user:123", {
  method: "POST",
  headers: {
    "Authorization": "YOUR_API_KEY",
    "Content-Type": "application/json"
  },
  body: JSON.stringify({
    value: {
      name: "John Doe",
      email: "john@example.com",
      roles: ["admin", "user"]
    }
  })
});
```

### Array

```javascript
await fetch("%%URL%%/my-plate/json/tags", {
  method: "POST",
  headers: {
    "Authorization": "YOUR_API_KEY",
    "Content-Type": "application/json"
  },
  body: JSON.stringify({
    value: ["red", "green", "blue"]
  })
});
```

### Primitive Values

```javascript
// Number
await fetch("%%URL%%/my-plate/json/count", {
  method: "POST",
  headers: {
    "Authorization": "YOUR_API_KEY",
    "Content-Type": "application/json"
  },
  body: JSON.stringify({
    value: 42
  })
});

// String
await fetch("%%URL%%/my-plate/json/greeting", {
  method: "POST",
  headers: {
    "Authorization": "YOUR_API_KEY",
    "Content-Type": "application/json"
  },
  body: JSON.stringify({
    value: "Hello World"
  })
});

// Boolean
await fetch("%%URL%%/my-plate/json/active", {
  method: "POST",
  headers: {
    "Authorization": "YOUR_API_KEY",
    "Content-Type": "application/json"
  },
  body: JSON.stringify({
    value: true
  })
});
```

### Nested Objects

```javascript
await fetch("%%URL%%/my-plate/json/order:123", {
  method: "POST",
  headers: {
    "Authorization": "YOUR_API_KEY",
    "Content-Type": "application/json"
  },
  body: JSON.stringify({
    value: {
      id: "123",
      customer: {
        name: "John",
        email: "john@example.com"
      },
      items: [
        { product: "Widget", quantity: 2, price: 9.99 }
      ],
      total: 19.98
    },
    ttl_ms: 86400000 // 24 hours
  })
});
```

---

## TTL Examples

| ttl_ms value | Duration |
|--------------|----------|
| 60000 | 1 minute |
| 3600000 | 1 hour |
| 86400000 | 1 day |
| 604800000 | 1 week |

```javascript
// 5 minutes
ttl_ms: 300000

// 30 minutes
ttl_ms: 1800000

// 1 day
ttl_ms: 86400000
```

---

## Error Responses

### Key Not Found (GET)

```json
{
  "error": "redis_nil",
  "message": "key does not exist"
}
```

### Invalid JSON (POST)

```json
{
  "error": "invalid_json",
  "message": "..."
}
```

---

## Use Cases

1. **User Profiles** - Store complete user objects
2. **Configuration** - Application settings
3. **Caching** - Store API responses
4. **Session Data** - User session information
5. **Documents** - Complex nested data structures
6. **Application State** - Store current state of an application

---

## Getting Started Example

```javascript
// Store a user's profile
async function saveUserProfile(userId, profile) {
  const response = await fetch(`%%URL%%/my-plate/json/user:${userId}`, {
    method: "POST",
    headers: {
      "Authorization": "YOUR_API_KEY",
      "Content-Type": "application/json"
    },
    body: JSON.stringify({
      value: profile,
      ttl_ms: 86400000 // 24 hours
    })
  });
  return response.json();
}

// Load a user's profile
async function getUserProfile(userId) {
  const response = await fetch(`%%URL%%/my-plate/json/user:${userId}`, {
    headers: { "Authorization": "YOUR_API_KEY" }
  });
  if (!response.ok) return null;
  return response.json();
}

// Usage
await saveUserProfile("123", {
  name: "John",
  email: "john@example.com",
  preferences: { theme: "dark" }
});

const profile = await getUserProfile("123");
console.log(profile.name); // "John"
```