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

## Command Compatibility

```json
POST /{plateID}/lists/command
{
  "command": "LPUSH",
  "args": ["mylist", "a", "b", "c"]
}
```
