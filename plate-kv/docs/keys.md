# Keys API

The Keys API handles inspection, scan operations, expiry, rename, copy, and deletion.

## Endpoints

| Method | Endpoint | Description |
| --- | --- | --- |
| GET | `/{plateID}/keys/{key}` | Get metadata and string preview |
| DELETE | `/{plateID}/keys/exact/{key}` | Delete one exact key |
| POST | `/{plateID}/keys/delete` | Delete many exact keys |
| POST | `/{plateID}/keys/{key}/expire` | Set TTL in milliseconds |
| DELETE | `/{plateID}/keys/{key}/expire` | Remove TTL |
| POST | `/{plateID}/keys/{key}/rename` | Rename a key |
| POST | `/{plateID}/keys/{key}/copy` | Copy a key |
| GET | `/{plateID}/scan` | Scan keys |
| POST | `/{plateID}/scan/hashes/{key}` | `HSCAN` |
| POST | `/{plateID}/scan/sets/{key}` | `SSCAN` |
| POST | `/{plateID}/scan/zsets/{key}` | `ZSCAN` |
| DELETE | `/{plateID}/keys/{pattern}` | Delete by pattern |
| POST | `/{plateID}/keys/command` | Execute allowed key commands |
| POST | `/{plateID}/keys/{key}/command` | Execute key-specific key commands |

## Inspection

```text
GET /{plateID}/keys/mykey
```

This returns metadata such as:

- `key`
- `exists`
- `type`
- `ttl_ms`
- `value` for string keys

`ttl_ms` follows Redis `PTTL` semantics:

- `-1` means the key exists but has no expiry.
- `-2` means the key does not exist.

## Exact Delete vs Pattern Delete

Exact delete:

```text
DELETE /{plateID}/keys/exact/mykey
```

Delete many exact keys:

```json
POST /{plateID}/keys/delete
{
  "keys": ["key1", "key2"]
}
```

Legacy pattern delete is still available:

```text
DELETE /{plateID}/keys/user:*
```

## Expiry, Rename, Copy

Set expiry:

```json
POST /{plateID}/keys/mykey/expire
{
  "ttl_ms": 60000
}
```

Remove expiry:

```text
DELETE /{plateID}/keys/mykey/expire
```

Rename:

```json
POST /{plateID}/keys/mykey/rename
{
  "destination": "mykey:new"
}
```

Copy:

```json
POST /{plateID}/keys/mykey/copy
{
  "destination": "mykey:copy",
  "replace": true
}
```

## Scan Examples

Scan keys:

```text
GET /{plateID}/scan?cursor=0&count=100&match=user:*
```

Scan a hash:

```json
POST /{plateID}/scan/hashes/user:1
{
  "cursor": 0,
  "match": "*",
  "count": 100
}
```
