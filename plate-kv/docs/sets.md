# Sets API

Sets store unique members with no guaranteed order. They are useful for tags, memberships, feature flags, and unique identity collections.

## Endpoints

| Method | Endpoint | Description |
| --- | --- | --- |
| POST | `/{plateID}/sets/add` | Add members |
| POST | `/{plateID}/sets/remove` | Remove members |
| GET | `/{plateID}/sets/members/{key}` | Read all members |
| POST | `/{plateID}/sets/contains` | Check one or many members |
| GET | `/{plateID}/sets/count/{key}` | Count members |
| GET | `/{plateID}/sets/random/{key}` | Read random members |
| POST | `/{plateID}/sets/pop` | Pop random members |
| POST | `/{plateID}/sets/union` | Union or union-store |
| POST | `/{plateID}/sets/intersect` | Intersection or store |
| POST | `/{plateID}/sets/diff` | Difference or store |
| POST | `/{plateID}/sets/command` | Execute allowed set commands |
| POST | `/{plateID}/sets/{key}/command` | Execute key-specific set commands |

## Examples

Add members:

```json
POST /{plateID}/sets/add
{
  "key": "tags",
  "members": ["red", "blue", "green"]
}
```

Check one member:

```json
POST /{plateID}/sets/contains
{
  "key": "tags",
  "member": "red"
}
```

Check many members:

```json
POST /{plateID}/sets/contains
{
  "key": "tags",
  "members": ["red", "black"]
}
```

Union without storing:

```json
POST /{plateID}/sets/union
{
  "keys": ["team:a", "team:b"]
}
```

Union and store the result:

```json
POST /{plateID}/sets/union
{
  "keys": ["team:a", "team:b"],
  "destination": "team:all"
}
```

## Command Compatibility

```json
POST /{plateID}/sets/command
{
  "command": "SADD",
  "args": ["myset", "a", "b", "c"]
}
```
