# Geo API

The Geo API provides operations for **geospatial data**. Perfect for storing locations and performing distance calculations!

## Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/{plateID}/geo/command` | Execute geo commands |
| POST | `/{plateID}/geo/{key}/command` | Execute key-specific geo commands |

---

## What is Geo?

The Geo commands let you:
- Store locations (latitude + longitude)
- Find distances between locations
- Search for locations within a radius

Under the hood, locations are stored as sorted sets with geohash scores.

---

## GEOADD - Add Locations

Add one or more locations with their coordinates.

**Request:**
```json
{
  "command": "GEOADD",
  "args": ["cities", "-122.4194", "37.7749", "San Francisco", "-118.2437", "34.0522", "Los Angeles"]
}
```

**Response:**
```json
{
  "result": 2
}
```
Number of locations added.

### JavaScript

```javascript
const response = await fetch("%%URL%%/my-plate/geo/command", {
  method: "POST",
  headers: {
    "Authorization": "YOUR_API_KEY",
    "Content-Type": "application/json"
  },
  body: JSON.stringify({
    command: "GEOADD",
    args: ["cities", -122.4194, 37.7749, "San Francisco", -118.2437, 34.0522, "Los Angeles"]
  })
});
// result: 2
```

### cURL

```bash
curl -X POST "%%URL%%/my-plate/geo/command" \
  -H "Authorization: YOUR_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{"command":"GEOADD","args":["cities","-122.4194","37.7749","San Francisco","-118.2437","34.0522","Los Angeles"]}'
```

### Format

Arguments: `longitude latitude name [longitude latitude name ...]`

- First longitude, then latitude!
- Example: San Francisco at (-122.4194, 37.7749)

---

## GEOPOS - Get Coordinates

Get the coordinates for one or more location names.

**Request:**
```json
{
  "command": "GEOPOS",
  "args": ["cities", "San Francisco", "Los Angeles"]
}
```

**Response:**
```json
{
  "result": [
    ["-122.41939999175072", "37.77439994509436"],
    ["-118.24369901405334", "34.05219998740545"]
  ]
}
```

### cURL

```bash
curl -X POST "%%URL%%/my-plate/geo/command" \
  -H "Authorization: YOUR_API_KEY" \
  -H "Content-Type": application/json" \
  -d '{"command":"GEOPOS","args":["cities","San Francisco","Los Angeles"]}'
```

---

## GEODIST - Calculate Distance

Calculate the distance between two locations.

**Request:**
```json
{
  "command": "GEODIST",
  "args": ["cities", "San Francisco", "Los Angeles", "km"]
}
```

**Response:**
```json
{
  "result": "559.2619"
}
```

### JavaScript

```javascript
const response = await fetch("%%URL%%/my-plate/geo/command", {
  method: "POST",
  headers: {
    "Authorization": "YOUR_API_KEY",
    "Content-Type": "application/json"
  },
  body: JSON.stringify({
    command: "GEODIST",
    args: ["cities", "San Francisco", "Los Angeles", "km"]
  })
});
// result: "559.2619"
```

### cURL

```bash
# Distance in kilometers
curl -X POST "%%URL%%/my-plate/geo/command" \
  -H "Authorization: YOUR_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{"command":"GEODIST","args":["cities","San Francisco","Los Angeles","km"]}'

# Distance in miles
curl -X POST "%%URL%%/my-plate/geo/command" \
  -H "Authorization: YOUR_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{"command":"GEODIST","args":["cities","San Francisco","Los Angeles","mi"]}'

# Distance in meters
curl -X POST "%%URL%%/my-plate/geo/command" \
  -H "Authorization: YOUR_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{"command":"GEODIST","args":["cities","San Francisco","Los Angeles","m"]}'
```

### Unit Options

- `m` - meters
- `km` - kilometers
- `mi` - miles
- `ft` - feet

---

## GEOSEARCH - Search by Radius

Find locations within a radius of a point.

**Request:**
```json
{
  "command": "GEOSEARCH",
  "args": ["cities", "FROMLONLAT", "-122.4194", "37.7749", "BYRADIUS", "100", "km"]
}
```

**Response:**
```json
{
  "result": ["San Francisco"]
}
```

### JavaScript

```javascript
const response = await fetch("%%URL%%/my-plate/geo/command", {
  method: "POST",
  headers: {
    "Authorization": "YOUR_API_KEY",
    "Content-Type": "application/json"
  },
  body: JSON.stringify({
    command: "GEOSEARCH",
    args: ["cities", "FROMLONLAT", -122.4194, 37.7749, "BYRADIUS", 100, "km"]
  })
});
// result: ["San Francisco"]
```

### cURL

```bash
# Find places within 100km of a point
curl -X POST "%%URL%%/my-plate/geo/command" \
  -H "Authorization: YOUR_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{"command":"GEOSEARCH","args":["cities","FROMLONLAT","-122.4194","37.7749","BYRADIUS","100","km"]}'
```

### GEOSEARCH with Options

Get results with distances and coordinates.

**Request:**
```json
{
  "command": "GEOSEARCH",
  "args": ["cities", "FROMLONLAT", "-122.4194", "37.7749", "BYRADIUS", "200", "km", "ASC", "WITHDIST", "WITHDOOR"]
}
```

**Response:**
```json
{
  "result": [
    ["San Francisco", "0.00", ["-122.419", "37.774"]],
    ["Los Angeles", "559.26", ["-118.243", "34.052"]]
  ]
}
```

### cURL

```bash
# With distance and coordinates
curl -X POST "%%URL%%/my-plate/geo/command" \
  -H "Authorization: YOUR_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{"command":"GEOSEARCH","args":["cities","FROMLONLAT","-122.4194","37.7749","BYRADIUS","200","km","ASC","WITHDIST","WITHDOOR"]}'
```

### Options

- `ASC` or `DESC` - sort order
- `WITHDIST` - include distance
- `WITHDOOR` - include coordinates
- `COUNT n` - limit results

---

## GEOSEARCHSTORE - Store Results

Search and store results to a new sorted set.

**Request:**
```json
{
  "command": "GEOSEARCHSTORE",
  "args": ["nearby", "cities", "FROMLONLAT", "-122.4194", "37.7749", "BYRADIUS", "100", "km"]
}
```

**Response:**
```json
{
  "result": 1
}
```

### cURL

```bash
curl -X POST "%%URL%%/my-plate/geo/command" \
  -H "Authorization: YOUR_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{"command":"GEOSEARCHSTORE","args":["nearby","cities","FROMLONLAT","-122.4194","37.7749","BYRADIUS","100","km"]}'
```

---

## Search by Box (Alternative)

You can also search within a rectangular box instead of radius.

```bash
curl -X POST "%%URL%%/my-plate/geo/command" \
  -H "Authorization: YOUR_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{"command":"GEOSEARCH","args":["cities","FROMLONLAT","-122.5","37.5","BYBOX","100","km","ASC"]}'
```

---

## Command Reference

| Command | Description | Example |
|---------|-------------|---------|
| GEOADD | Add locations | `["key", "lon", "lat", "member", ...]` |
| GEOPOS | Get positions | `["key", "member1", ...]` |
| GEODIST | Get distance | `["key", "m1", "m2", "unit"]` |
| GEOSEARCH | Search | `["key", "FROMLONLAT", "lon", "lat", "BYRADIUS", "r", "unit"]` |
| GEOSEARCHSTORE | Store search | `["dest", "key", ...]` |