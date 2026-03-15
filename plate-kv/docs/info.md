# Info API

The Info API provides statistics and information about your plate - useful for monitoring and debugging!

## Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/{plateID}/info` | Get plate statistics |

---

## What Does Info Show?

The Info endpoint gives you:
- **keyCount**: Total number of keys in your plate
- **memoryUsageBytes**: Estimated memory used by your data

This helps you:
- Monitor storage usage
- Plan capacity
- Debug performance issues

---

## GET /{plateID}/info

Get key count and memory usage for your plate.

### JavaScript

```javascript
const response = await fetch("%%URL%%/my-plate/info", {
  headers: {
    "Authorization": "YOUR_API_KEY"
  }
});
const info = await response.json();
// { keyCount: 150, memoryUsageBytes: 5242880 }
```

### cURL

```bash
curl -X GET "%%URL%%/my-plate/info" \
  -H "Authorization: YOUR_API_KEY"
```

---

## Response Fields

| Field | Type | Description |
|-------|------|-------------|
| keyCount | number | Total number of keys in the plate |
| memoryUsageBytes | number | Estimated memory usage in bytes |

### Example Output

```json
{
  "keyCount": 150,
  "memoryUsageBytes": 5242880
}
```
