# Sorted Sets API

Sorted sets combine unique members with numeric scores, making them a good fit for leaderboards, ranking, priorities, and time-based indexes.

## Endpoints

| Method | Endpoint | Description |
| --- | --- | --- |
| POST | `/{plateID}/zsets/add` | Add scored members |
| POST | `/{plateID}/zsets/remove` | Remove members |
| GET | `/{plateID}/zsets/score/{key}/{member}` | Get a score |
| GET | `/{plateID}/zsets/rank/{key}/{member}` | Get a rank |
| GET | `/{plateID}/zsets/range/{key}` | Read a range |
| GET | `/{plateID}/zsets/count/{key}` | Count members |
| GET | `/{plateID}/zsets/count-by-score/{key}` | Count by score window |
| GET | `/{plateID}/zsets/count-by-lex/{key}` | Count by lex window |
| POST | `/{plateID}/zsets/increment` | Increment a score |
| POST | `/{plateID}/zsets/pop-min` | Pop lowest member(s) |
| POST | `/{plateID}/zsets/pop-max` | Pop highest member(s) |
| GET | `/{plateID}/zsets/random/{key}` | Read random member(s) |
| POST | `/{plateID}/zsets/scores` | Get multiple scores |
| POST | `/{plateID}/zsets/union` | Union-store |
| POST | `/{plateID}/zsets/intersect` | Intersect-store |
| POST | `/{plateID}/zsets/diff` | Diff-store |
| POST | `/{plateID}/zsets/range/store` | `ZRANGESTORE` |
| POST | `/{plateID}/zsets/command` | Execute allowed zset commands |
| POST | `/{plateID}/zsets/{key}/command` | Execute key-specific zset commands |

## Examples

Add members:

```json
POST /{plateID}/zsets/add
{
  "key": "leaderboard",
  "members": [
    { "member": "player1", "score": 100 },
    { "member": "player2", "score": 200 }
  ]
}
```

Read a range with scores:

```text
GET /{plateID}/zsets/range/leaderboard?start=0&stop=9&with_scores=true
```

Read reverse rank:

```text
GET /{plateID}/zsets/rank/leaderboard/player2?order=desc
```

Increment a score:

```json
POST /{plateID}/zsets/increment
{
  "key": "leaderboard",
  "member": "player1",
  "amount": 10
}
```

Store a weighted union:

```json
POST /{plateID}/zsets/union
{
  "destination": "combined",
  "keys": ["set:a", "set:b"],
  "weights": [1, 2],
  "aggregate": "sum"
}
```

## Command Compatibility

```json
POST /{plateID}/zsets/command
{
  "command": "ZADD",
  "args": ["leaderboard", 100, "player1", 200, "player2"]
}
```
