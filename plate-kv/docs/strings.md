# Strings API

The Strings API stores single values per key. It is the simplest way to keep text, counters, tokens, feature flags, and cached serialized data.

## Endpoints

| Method | Endpoint | Description |
| --- | --- | --- |
| GET | `/{plateID}/strings/get/{key}` | Get one value |
| POST | `/{plateID}/strings/get` | Get many values |
| POST | `/{plateID}/strings/set` | Set one value |
| POST | `/{plateID}/strings/set-many` | Set many values |
| POST | `/{plateID}/strings/increment` | Increment integer or float |
| POST | `/{plateID}/strings/decrement` | Decrement integer |
| POST | `/{plateID}/strings/append` | Append to a string |
| GET | `/{plateID}/strings/length/{key}` | Get string length |
| GET | `/{plateID}/strings/range/{key}` | Read a substring |
| POST | `/{plateID}/strings/range/set` | Replace part of a string |
| POST | `/{plateID}/strings/get-and-expire` | `GETEX` wrapper |
| DELETE | `/{plateID}/strings/get-and-delete/{key}` | `GETDEL` wrapper |
| POST | `/{plateID}/strings/command` | Execute allowed string commands |
| POST | `/{plateID}/strings/{key}/command` | Execute key-specific string commands |

## Common Patterns

### Store a Value

```json
POST /{plateID}/strings/set
{
  "key": "mykey",
  "value": "hello world"
}
```

### Store with Expiration

```json
POST /{plateID}/strings/set
{
  "key": "session:123",
  "value": "abc123",
  "ttl_ms": 3600000
}
```

### Conditional Set

```json
POST /{plateID}/strings/set
{
  "key": "lock:job:42",
  "value": "worker-1",
  "nx": true,
  "ttl_ms": 30000
}
```

### Get a Value

```text
GET /{plateID}/strings/get/mykey
```

### Get Multiple Values

```json
POST /{plateID}/strings/get
{
  "keys": ["key1", "key2", "key3"]
}
```

### Set Multiple Values

```json
POST /{plateID}/strings/set-many
{
  "values": {
    "key1": "value1",
    "key2": "value2"
  }
}
```

### Increment Counters

```json
POST /{plateID}/strings/increment
{
  "key": "counter",
  "amount": 10
}
```

Float increments are supported too:

```json
POST /{plateID}/strings/increment
{
  "key": "price",
  "amount": 1.5
}
```

### Decrement

```json
POST /{plateID}/strings/decrement
{
  "key": "counter",
  "amount": 5
}
```

### Append

```json
POST /{plateID}/strings/append
{
  "key": "greeting",
  "value": " world"
}
```

### Range Operations

Read a substring:

```text
GET /{plateID}/strings/range/greeting?start=0&end=4
```

Query parameters:
- `start` (optional): Start index (default: 0)
- `end` (optional): End index (default: -1, meaning to the end)

Replace part of a string:

```json
POST /{plateID}/strings/range/set
{
  "key": "greeting",
  "offset": 6,
  "value": "world"
}
```

### Get and Expire

```json
POST /{plateID}/strings/get-and-expire
{
  "key": "session:123",
  "ttl_ms": 60000
}
```

Set `ttl_ms` to `0` to persist the key while returning the value.

### Get and Delete

```text
DELETE /{plateID}/strings/get-and-delete/session:123
```

## Command Compatibility

If you need direct command access, use:

```json
POST /{plateID}/strings/command
{
  "command": "SET",
  "args": ["mykey", "hello world", "EX", "3600"]
}
```

## Command Endpoints

### `POST /{plateID}/strings/command`

Execute allowed string commands across the plate.

**Allowed Commands:**

| Command | Description |
|---------|-------------|
| GET | Get value |
| SET | Set value |
| MGET | Get multiple values |
| MSET | Set multiple values |
| GETEX | Get and optionally set expiry |
| GETDEL | Get and delete |
| INCR | Increment by 1 |
| DECR | Decrement by 1 |
| INCRBY | Increment by integer |
| DECRBY | Decrement by integer |
| INCRBYFLOAT | Increment by float |
| APPEND | Append to string |
| STRLEN | Get string length |
| SETRANGE | Replace substring |
| GETRANGE | Get substring |

**Request:**

```json
POST /{plateID}/strings/command
{
  "command": "INCR",
  "args": ["counter"]
}
```

### `POST /{plateID}/strings/{key}/command`

Execute allowed commands on a specific string key.

**Allowed Commands:** Same as above.

**Request:**

```json
POST /{plateID}/strings/mykey/command
{
  "command": "STRLEN"
}
```
