# Sorted Sets API

The Sorted Sets API (also called ZSets) provides operations for **sorted set data types**. Each member has a score, and the set is sorted by score - like a leaderboard!

## Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/{plateID}/zsets/command` | Execute sorted set commands |
| POST | `/{plateID}/zsets/{key}/command` | Execute key-specific sorted set commands |

---

## What is a Sorted Set?

A sorted set combines:
- **Uniqueness** of a set (no duplicates)
- **Ordering** by a numeric score

Use cases:
- Leaderboards (highest score at top)
- Priority queues
- Time-series data (timestamps as scores)

---

## ZADD - Add Members with Scores

Add members with their scores to the sorted set.

**Request:**
```json
{
  "command": "ZADD",
  "args": ["leaderboard", "100", "player1", "200", "player2", "150", "player3"]
}
```

**Response:**
```json
{
  "result": 3
}
```
Number of new members added.

### JavaScript

```javascript
const response = await fetch("%%URL%%/my-plate/zsets/command", {
  method: "POST",
  headers: {
    "Authorization": "YOUR_API_KEY",
    "Content-Type": "application/json"
  },
  body: JSON.stringify({
    command: "ZADD",
    args: ["leaderboard", 100, "player1", 200, "player2", 150, "player3"]
  })
});
// result: 3
```

### cURL

```bash
curl -X POST "%%URL%%/my-plate/zsets/command" \
  -H "Authorization: YOUR_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{"command":"ZADD","args":["leaderboard","100","player1","200","player2","150","player3"]}'
```

### Result (sorted by score):

| Rank | Member | Score |
|------|--------|-------|
| 1 | player1 | 100 |
| 2 | player3 | 150 |
| 3 | player2 | 200 |

---

## ZREM - Remove Members

Remove one or more members.

**Request:**
```json
{
  "command": "ZREM",
  "args": ["leaderboard", "player1"]
}
```

**Response:**
```json
{
  "result": 1
}
```
Number of members removed.

### cURL

```bash
curl -X POST "%%URL%%/my-plate/zsets/command" \
  -H "Authorization: YOUR_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{"command":"ZREM","args":["leaderboard","player1"]}'
```

---

## ZSCORE - Get Score

Get the score of a specific member.

**Request:**
```json
{
  "command": "ZSCORE",
  "args": ["leaderboard", "player2"]
}
```

**Response:**
```json
{
  "result": "200"
}
```

### cURL

```bash
curl -X POST "%%URL%%/my-plate/zsets/command" \
  -H "Authorization: YOUR_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{"command":"ZSCORE","args":["leaderboard","player2"]}'
```

---

## ZRANK - Get Rank (Ascending)

Get the position of a member when sorted from lowest to highest score.

- Rank 0 = lowest score

**Request:**
```json
{
  "command": "ZRANK",
  "args": ["leaderboard", "player2"]
}
```

**Response:**
```json
{
  "result": 2
}
```

### JavaScript

```javascript
const response = await fetch("%%URL%%/my-plate/zsets/command", {
  method: "POST",
  headers: {
    "Authorization": "YOUR_API_KEY",
    "Content-Type": "application/json"
  },
  body: JSON.stringify({
    command: "ZRANK",
    args: ["leaderboard", "player2"]
  })
});
// result: 2
```

### cURL

```bash
curl -X POST "%%URL%%/my-plate/zsets/command" \
  -H "Authorization: YOUR_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{"command":"ZRANK","args":["leaderboard","player2"]}'
```

---

## ZREVRANK - Get Rank (Descending)

Get the position when sorted from highest to lowest score.

- Rank 0 = highest score

**Request:**
```json
{
  "command": "ZREVRANK",
  "args": ["leaderboard", "player2"]
}
```

**Response:**
```json
{
  "result": 0
}
```

### cURL

```bash
curl -X POST "%%URL%%/my-plate/zsets/command" \
  -H "Authorization: YOUR_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{"command":"ZREVRANK","args":["leaderboard","player2"]}'
```

---

## ZRANGE - Get Members by Rank

Get members in a rank range (low to high).

**Request:**
```json
{
  "command": "ZRANGE",
  "args": ["leaderboard", "0", "2", "WITHSCORES"]
}
```

**Response:**
```json
{
  "result": ["player3", "150", "player2", "200"]
}
```

### JavaScript

```javascript
const response = await fetch("%%URL%%/my-plate/zsets/command", {
  method: "POST",
  headers: {
    "Authorization": "YOUR_API_KEY",
    "Content-Type": "application/json"
  },
  body: JSON.stringify({
    command: "ZRANGE",
    args: ["leaderboard", 0, 2, "WITHSCORES"]
  })
});
// result: ["player3", "150", "player2", "200"]
```

### cURL

```bash
# Get top 3 with scores
curl -X POST "%%URL%%/my-plate/zsets/command" \
  -H "Authorization: YOUR_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{"command":"ZRANGE","args":["leaderboard","0","2","WITHSCORES"]}'

# Get top 3 without scores
curl -X POST "%%URL%%/my-plate/zsets/command" \
  -H "Authorization: YOUR_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{"command":"ZRANGE","args":["leaderboard","0","2"]}'
```

---

## ZCARD - Get Count

Get the number of members.

**Request:**
```json
{
  "command": "ZCARD",
  "args": ["leaderboard"]
}
```

**Response:**
```json
{
  "result": 3
}
```

### cURL

```bash
curl -X POST "%%URL%%/my-plate/zsets/command" \
  -H "Authorization: YOUR_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{"command":"ZCARD","args":["leaderboard"]}'
```

---

## ZCOUNT - Count by Score Range

Count members with scores in a range.

**Request:**
```json
{
  "command": "ZCOUNT",
  "args": ["leaderboard", "100", "200"]
}
```

**Response:**
```json
{
  "result": 3
}
```

### cURL

```bash
curl -X POST "%%URL%%/my-plate/zsets/command" \
  -H "Authorization: YOUR_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{"command":"ZCOUNT","args":["leaderboard","100","200"]}'
```

---

## ZLEXCOUNT - Count by Lexicographical Range

Count members within a lexicographical range (when scores are equal).

**Request:**
```json
{
  "command": "ZLEXCOUNT",
  "args": ["myzset", "[a", "[z"]
}
```

**Response:**
```json
{
  "result": 10
}
```

### cURL

```bash
curl -X POST "%%URL%%/my-plate/zsets/command" \
  -H "Authorization: YOUR_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{"command":"ZLEXCOUNT","args":["myzset","[a","[z"]}'
```

---

## ZINCRBY - Increment Score

Increase a member's score by a value.

**Request:**
```json
{
  "command": "ZINCRBY",
  "args": ["leaderboard", "50", "player3"]
}
```

**Response:**
```json
{
  "result": "200"
}
```
New score is returned.

### JavaScript

```javascript
const response = await fetch("%%URL%%/my-plate/zsets/command", {
  method: "POST",
  headers: {
    "Authorization": "YOUR_API_KEY",
    "Content-Type": "application/json"
  },
  body: JSON.stringify({
    command: "ZINCRBY",
    args: ["leaderboard", 50, "player3"]
  })
});
// result: "200"
```

### cURL

```bash
curl -X POST "%%URL%%/my-plate/zsets/command" \
  -H "Authorization: YOUR_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{"command":"ZINCRBY","args":["leaderboard","50","player3"]}'
```

### Use Case: Update Leaderboard

```javascript
// Add points to a player's score
async function addPoints(playerId, points) {
  await fetch("%%URL%%/my-plate/zsets/command", {
    method: "POST",
    headers: {
      "Authorization": "YOUR_API_KEY",
      "Content-Type": "application/json"
    },
    body: JSON.stringify({
      command: "ZINCRBY",
      args: ["leaderboard", points, playerId]
    })
  });
}

// Player scored 10 more points
addPoints("player1", 10);
```

---

## ZPOPMIN - Pop Lowest Score Members

Remove and return members with lowest scores.

**Request:**
```json
{
  "command": "ZPOPMIN",
  "args": ["leaderboard", "2"]
}
```

**Response:**
```json
{
  "result": ["player3", "150"]
}
```

### cURL

```bash
curl -X POST "%%URL%%/my-plate/zsets/command" \
  -H "Authorization: YOUR_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{"command":"ZPOPMIN","args":["leaderboard","2"]}'
```

---

## ZPOPMAX - Pop Highest Score Members

Remove and return members with highest scores.

**Request:**
```json
{
  "command": "ZPOPMAX",
  "args": ["leaderboard", "2"]
}
```

**Response:**
```json
{
  "result": ["player2", "200"]
}
```

### cURL

```bash
curl -X POST "%%URL%%/my-plate/zsets/command" \
  -H "Authorization: YOUR_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{"command":"ZPOPMAX","args":["leaderboard","2"]}'
```

---

## ZRANDMEMBER - Random Members

Get random member(s).

**Request:**
```json
{
  "command": "ZRANDMEMBER",
  "args": ["leaderboard", "2", "WITHSCORES"]
}
```

**Response:**
```json
{
  "result": ["player1", "100", "player3", "150"]
}
```

### cURL

```bash
curl -X POST "%%URL%%/my-plate/zsets/command" \
  -H "Authorization: YOUR_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{"command":"ZRANDMEMBER","args":["leaderboard","2","WITHSCORES"]}'
```

---

## Store Operations

### ZUNIONSTORE - Combine Multiple Sets

```bash
curl -X POST "%%URL%%/my-plate/zsets/command" \
  -H "Authorization: YOUR_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{"command":"ZUNIONSTORE","args":["dest","2","set1","set2"]}'
```

### ZINTERSTORE - Intersect Sets

```bash
curl -X POST "%%URL%%/my-plate/zsets/command" \
  -H "Authorization: YOUR_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{"command":"ZINTERSTORE","args":["dest","2","set1","set2"]}'
```

---

## ZMSCORE - Get Multiple Scores

Get scores for multiple members at once.

**Request:**
```json
{
  "command": "ZMSCORE",
  "args": ["leaderboard", "player1", "player2"]
}
```

**Response:**
```json
{
  "result": ["100", "200"]
}
```

### cURL

```bash
curl -X POST "%%URL%%/my-plate/zsets/command" \
  -H "Authorization: YOUR_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{"command":"ZMSCORE","args":["leaderboard","player1","player2"]}'
```

---

## Command Reference

| Command | Description | Example |
|---------|-------------|---------|
| ZADD | Add members | `["key", "score", "member", ...]` |
| ZREM | Remove members | `["key", "member1", ...]` |
| ZSCORE | Get score | `["key", "member"]` |
| ZRANK | Get rank (asc) | `["key", "member"]` |
| ZREVRANK | Get rank (desc) | `["key", "member"]` |
| ZRANGE | Get range | `["key", "start", "stop", "WITHSCORES"]` |
| ZCARD | Get count | `["key"]` |
| ZCOUNT | Count by score | `["key", "min", "max"]` |
| ZLEXCOUNT | Count by lex | `["key", "min", "max"]` |
| ZINCRBY | Incr score | `["key", "incr", "member"]` |
| ZPOPMIN | Pop min | `["key", "count"]` |
| ZPOPMAX | Pop max | `["key", "count"]` |
| ZRANDMEMBER | Random | `["key", "count", "WITHSCORES"]` |
| ZUNIONSTORE | Store union | `["dest", "num", "key1", "key2"]` |
| ZINTERSTORE | Store inter | `["dest", "num", "key1", "key2"]` |
| ZMSCORE | Multi score | `["key", "m1", "m2", ...]` |