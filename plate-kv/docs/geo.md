# Geo API

Geo endpoints store named locations and perform distance and area searches. Under the hood, Redis stores geospatial members in sorted sets.

## Endpoints

| Method | Endpoint | Description |
| --- | --- | --- |
| POST | `/{plateID}/geo/add` | Add locations |
| POST | `/{plateID}/geo/positions` | Get coordinates |
| GET | `/{plateID}/geo/distance/{key}` | Get distance between members |
| POST | `/{plateID}/geo/search` | `GEOSEARCH` |
| POST | `/{plateID}/geo/search/store` | `GEOSEARCHSTORE` |
| POST | `/{plateID}/geo/command` | Execute allowed geo commands |
| POST | `/{plateID}/geo/{key}/command` | Execute key-specific geo commands |

## Examples

Add locations:

```json
POST /{plateID}/geo/add
{
  "key": "cities",
  "locations": [
    { "member": "San Francisco", "longitude": -122.4194, "latitude": 37.7749 },
    { "member": "Los Angeles", "longitude": -118.2437, "latitude": 34.0522 }
  ]
}
```

Get positions:

```json
POST /{plateID}/geo/positions
{
  "key": "cities",
  "members": ["San Francisco", "Los Angeles"]
}
```

Get distance:

```text
GET /{plateID}/geo/distance/cities?from=San%20Francisco&to=Los%20Angeles&unit=km
```

Search by radius:

```json
POST /{plateID}/geo/search
{
  "key": "cities",
  "from_member": "San Francisco",
  "radius": 600,
  "unit": "km",
  "with_dist": true,
  "sort": "asc"
}
```

Search by box and store the result:

```json
POST /{plateID}/geo/search/store
{
  "key": "cities",
  "destination": "cities:west-coast",
  "from_lon": -122.4194,
  "from_lat": 37.7749,
  "width": 800,
  "height": 800,
  "unit": "km"
}
```

## Command Compatibility

```json
POST /{plateID}/geo/command
{
  "command": "GEOADD",
  "args": ["cities", -122.4194, 37.7749, "San Francisco"]
}
```

## Command Endpoints

### `POST /{plateID}/geo/command`

Execute allowed geo commands across the plate.

**Allowed Commands:**

| Command | Description |
|---------|-------------|
| GEOADD | Add locations |
| GEOPOS | Get coordinates |
| GEODIST | Get distance |
| GEOSEARCH | Search locations |
| GEOSEARCHSTORE | Search and store results |

**Query Parameters for GEODIST:**
- `from` (required): First member
- `to` (required): Second member
- `unit` (optional): Distance unit - `m` (meters), `km` (kilometers), `mi` (miles), `ft` (feet). Default: `m`

### `POST /{plateID}/geo/{key}/command`

Execute allowed commands on a specific geo key.

**Allowed Commands:** Same as above.

**Request:**

```json
POST /{plateID}/geo/cities/command
{
  "command": "GEOPOS",
  "args": ["San Francisco"]
}
```
