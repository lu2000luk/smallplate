# Bitmaps API

Bitmaps are efficient for yes/no style data. Each bit position stores `0` or `1`, which makes them useful for activity tracking, feature flags, and compact boolean datasets.

## Endpoints

| Method | Endpoint | Description |
| --- | --- | --- |
| POST | `/{plateID}/bitmaps/set` | Set one bit |
| GET | `/{plateID}/bitmaps/get/{key}/{bit}` | Get one bit |
| GET | `/{plateID}/bitmaps/count/{key}` | Count set bits |
| GET | `/{plateID}/bitmaps/position/{key}/{bit}` | Find first matching bit |
| POST | `/{plateID}/bitmaps/operate` | `BITOP` wrapper |
| POST | `/{plateID}/bitmaps/field` | `BITFIELD` wrapper |
| POST | `/{plateID}/bitmaps/command` | Execute allowed bitmap commands |
| POST | `/{plateID}/bitmaps/{key}/command` | Execute key-specific bitmap commands |

## Examples

### Set a Bit

```json
POST /{plateID}/bitmaps/set
{
  "key": "login:2026-03-15",
  "bit": 12345,
  "value": 1
}
```

### Get a Bit

```text
GET /{plateID}/bitmaps/get/login:2026-03-15/12345
```

### Count Bits

Count the whole bitmap:

```text
GET /{plateID}/bitmaps/count/login:2026-03-15
```

Count a byte range:

```text
GET /{plateID}/bitmaps/count/login:2026-03-15?start=0&end=100
```

### Find the First `0` or `1`

```text
GET /{plateID}/bitmaps/position/login:2026-03-15/0?start=0&end=100
```

### Bitwise Operations

```json
POST /{plateID}/bitmaps/operate
{
  "operation": "AND",
  "destination": "both-days",
  "sources": ["day-1", "day-2"]
}
```

`NOT` requires exactly one source key.

### Advanced Bitfield Operations

```json
POST /{plateID}/bitmaps/field
{
  "key": "flags",
  "operations": ["GET", "u8", 0, "SET", "u8", 8, 255]
}
```

## Command Compatibility

```json
POST /{plateID}/bitmaps/command
{
  "command": "SETBIT",
  "args": ["mybitmap", 500, 1]
}
```

## Command Endpoints

### `POST /{plateID}/bitmaps/command`

Execute allowed bitmap commands across the plate.

**Allowed Commands:**

| Command | Description |
|---------|-------------|
| SETBIT | Set a bit |
| GETBIT | Get a bit |
| BITCOUNT | Count set bits |
| BITOP | Bitwise operations |
| BITPOS | Find first bit |
| BITFIELD | Bitfield operations |

### `POST /{plateID}/bitmaps/{key}/command`

Execute allowed commands on a specific bitmap key.

**Allowed Commands:** Same as above.

**Request:**

```json
POST /{plateID}/bitmaps/mybitmap/command
{
  "command": "BITCOUNT"
}
```
