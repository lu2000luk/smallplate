# Hashes API

Hashes are ideal for object-like data. A single key contains multiple field-value pairs, which makes hashes a great fit for user profiles, counters, metadata, and cached records.

## Endpoints

| Method | Endpoint | Description |
| --- | --- | --- |
| POST | `/{plateID}/hashes/set` | Set one field or many fields |
| GET | `/{plateID}/hashes/get/{key}` | Get all fields |
| GET | `/{plateID}/hashes/get/{key}/{field}` | Get one field |
| POST | `/{plateID}/hashes/get` | Get many fields |
| DELETE | `/{plateID}/hashes/delete/{key}/{field}` | Delete one field |
| DELETE | `/{plateID}/hashes/delete/{key}?allow_delete_key=true` | Delete the full hash key |
| POST | `/{plateID}/hashes/delete` | Delete many fields |
| GET | `/{plateID}/hashes/exists/{key}/{field}` | Check field existence |
| GET | `/{plateID}/hashes/keys/{key}` | List fields |
| GET | `/{plateID}/hashes/values/{key}` | List values |
| GET | `/{plateID}/hashes/length/{key}` | Count fields |
| POST | `/{plateID}/hashes/increment` | Increment numeric field |
| POST | `/{plateID}/hashes/set-if-absent` | `HSETNX` wrapper |
| GET | `/{plateID}/hashes/random/{key}` | Random field(s) |
| POST | `/{plateID}/hashes/command` | Execute allowed hash commands |
| POST | `/{plateID}/hashes/{key}/command` | Execute key-specific hash commands |

## What Gets Stored

### Set a Single Field

```json
POST /{plateID}/hashes/set
{
  "key": "user:1",
  "field": "name",
  "value": "Ada"
}
```

### Set Many Fields from an Object

```json
POST /{plateID}/hashes/set
{
  "key": "user:1",
  "value": {
    "name": "Ada",
    "age": 30,
    "email": "ada@example.com"
  }
}
```

### Set Many Fields from a Stringified JSON Object

```json
POST /{plateID}/hashes/set
{
  "key": "user:1",
  "value": "{\"name\":\"Ada\",\"age\":30}"
}
```

Top-level object keys become hash fields automatically. Nested objects and arrays are JSON-stringified per field instead of flattened.

## Reads

### Get One Field

```text
GET /{plateID}/hashes/get/user:1/name
```

### Get All Fields

```text
GET /{plateID}/hashes/get/user:1
```

### Get Multiple Fields

```json
POST /{plateID}/hashes/get
{
  "key": "user:1",
  "fields": ["name", "age", "email"]
}
```

### Check Existence

```text
GET /{plateID}/hashes/exists/user:1/name
```

## Delete Safety

Deleting fields is the default:

```text
DELETE /{plateID}/hashes/delete/user:1/age
```

Delete many fields:

```json
POST /{plateID}/hashes/delete
{
  "key": "user:1",
  "fields": ["age", "email"]
}
```

Deleting the whole hash key is intentionally blocked unless you opt in:

```text
DELETE /{plateID}/hashes/delete/user:1?allow_delete_key=true
```

## Numeric Updates

```json
POST /{plateID}/hashes/increment
{
  "key": "user:1",
  "field": "points",
  "amount": 5
}
```

Floating-point amounts automatically use `HINCRBYFLOAT`.

## Other Examples

Set if absent:

```json
POST /{plateID}/hashes/set-if-absent
{
  "key": "user:1",
  "field": "nickname",
  "value": "A"
}
```

Random fields:

```text
GET /{plateID}/hashes/random/user:1?count=2&with_values=true
```

## Command Compatibility

```json
POST /{plateID}/hashes/command
{
  "command": "HSET",
  "args": ["user:1", "name", "Ada", "age", "30"]
}
```

## Command Endpoints

### `POST /{plateID}/hashes/command`

Execute allowed hash commands across the plate.

**Allowed Commands:**

| Command | Description |
|---------|-------------|
| HSET | Set field(s) |
| HGET | Get one field |
| HMGET | Get multiple fields |
| HGETALL | Get all fields |
| HDEL | Delete field(s) |
| HEXISTS | Check field existence |
| HINCRBY | Increment integer |
| HINCRBYFLOAT | Increment float |
| HKEYS | Get all field names |
| HVALS | Get all values |
| HLEN | Get field count |
| HSETNX | Set if absent |
| HRANDFIELD | Get random field(s) |

**Query Parameters for HRANDFIELD:**
- `count` (optional): Number of fields to return (negative for distinct)
- `with_values` (optional): Include values in response

### `POST /{plateID}/hashes/{key}/command`

Execute allowed commands on a specific hash key.

**Allowed Commands:** Same as above.

**Request:**

```json
POST /{plateID}/hashes/user:1/command
{
  "command": "HLEN"
}
```
