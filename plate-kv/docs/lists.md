# Lists API

Lists are ordered collections. They are useful for queues, stacks, feeds, and ordered work items.

## Endpoints

| Method | Endpoint | Description |
| --- | --- | --- |
| POST | `/{plateID}/lists/left/push` | Push to the left |
| POST | `/{plateID}/lists/right/push` | Push to the right |
| POST | `/{plateID}/lists/left/pop` | Pop from the left |
| POST | `/{plateID}/lists/right/pop` | Pop from the right |
| GET | `/{plateID}/lists/range/{key}` | Read a range |
| GET | `/{plateID}/lists/item/{key}/{index}` | Read one item |
| GET | `/{plateID}/lists/length/{key}` | Count items |
| POST | `/{plateID}/lists/position` | Find an item's position |
| POST | `/{plateID}/lists/set` | Replace an item by index |
| POST | `/{plateID}/lists/insert` | Insert before or after a pivot |
| POST | `/{plateID}/lists/remove` | Remove matching items |
| POST | `/{plateID}/lists/trim` | Trim list bounds |
| POST | `/{plateID}/lists/move` | Move between lists |
| POST | `/{plateID}/lists/command` | Execute allowed list commands |
| POST | `/{plateID}/lists/{key}/command` | Execute key-specific list commands |

## Examples

Push to the right:

```json
POST /{plateID}/lists/right/push
{
  "key": "jobs",
  "values": ["a", "b", "c"]
}
```

Pop two items from the left:

```json
POST /{plateID}/lists/left/pop
{
  "key": "jobs",
  "count": 2
}
```

Read a window:

```text
GET /{plateID}/lists/range/jobs?start=0&stop=9
```

Insert relative to a pivot:

```json
POST /{plateID}/lists/insert
{
  "key": "jobs",
  "where": "before",
  "pivot": "b",
  "value": "urgent"
}
```

Move an item between lists:

```json
POST /{plateID}/lists/move
{
  "source": "pending",
  "destination": "processing",
  "from": "right",
  "to": "left"
}
```

### Get Item by Index

```text
GET /{plateID}/lists/item/jobs/0
```

### Get Length

```text
GET /{plateID}/lists/length/jobs
```

### Find Position (LPOS)

```json
POST /{plateID}/lists/position
{
  "key": "jobs",
  "value": "urgent",
  "rank": 1,
  "count": 5
}
```

Query parameters (via JSON body):
- `rank` (optional): Start searching from this position
- `count` (optional): Number of positions to return
- `maxlen` (optional): Limit search depth

### Set Item by Index

```json
POST /{plateID}/lists/set
{
  "key": "jobs",
  "index": 0,
  "value": "replaced"
}
```

### Remove Items

```json
POST /{plateID}/lists/remove
{
  "key": "jobs",
  "value": "done",
  "count": 1
}
```

Set `count` to 0 to remove all matching items.

### Trim List

```json
POST /{plateID}/lists/trim
{
  "key": "jobs",
  "start": 0,
  "stop": 99
}
```

Keeps only elements from index 0 to 99, removing everything after.

## Command Compatibility

```json
POST /{plateID}/lists/command
{
  "command": "LPUSH",
  "args": ["mylist", "a", "b", "c"]
}
```

## Command Endpoints

### `POST /{plateID}/lists/command`

Execute allowed list commands across the plate.

**Allowed Commands:**

| Command | Description |
|---------|-------------|
| LPUSH | Push to left |
| RPUSH | Push to right |
| LPOP | Pop from left |
| RPOP | Pop from right |
| LLEN | Get length |
| LRANGE | Get range |
| LINDEX | Get by index |
| LPOS | Find position |
| LSET | Set by index |
| LINSERT | Insert relative to pivot |
| LREM | Remove items |
| LTRIM | Trim list |
| LMOVE | Move between lists |

**Query Parameters for LRANGE:**
- `start` (optional): Start index (default: 0)
- `stop` (optional): End index (default: -1)

**Query Parameters for LPOS:**
- `rank` (optional): Start position from this rank
- `count` (optional): Number of matches to return
- `maxlen` (optional): Limit search length

### `POST /{plateID}/lists/{key}/command`

Execute allowed commands on a specific list key.

**Allowed Commands:** Same as above.

**Request:**

```json
POST /{plateID}/lists/mylist/command
{
  "command": "LLEN"
}
```
