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

## Command Endpoints

### `POST /{plateID}/sets/command`

Execute allowed set commands across the plate.

**Allowed Commands:**

| Command | Description |
|---------|-------------|
| SADD | Add members |
| SREM | Remove members |
| SMEMBERS | Get all members |
| SISMEMBER | Check one member |
| SMISMEMBER | Check multiple members |
| SCARD | Get member count |
| SRANDMEMBER | Get random members |
| SPOP | Pop random members |
| SUNION | Union of sets |
| SINTER | Intersection of sets |
| SDIFF | Difference of sets |
| SUNIONSTORE | Union and store |
| SINTERSTORE | Intersection and store |
| SDIFFSTORE | Difference and store |

**Query Parameters for SRANDMEMBER:**
- `count` (optional): Number of members to return (negative for distinct)

### `POST /{plateID}/sets/{key}/command`

Execute allowed commands on a specific set key.

**Allowed Commands:** Same as above.

**Request:**

```json
POST /{plateID}/sets/myset/command
{
  "command": "SCARD"
}
```
