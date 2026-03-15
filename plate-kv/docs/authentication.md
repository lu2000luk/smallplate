# Authentication

All API endpoints require authentication using the `Authorization` header with your API key.

## Header Format

```
Authorization: YOUR_API_KEY
```

## Examples

### JavaScript (fetch)

```javascript
const plateID = "my-plate";
const apiKey = "your-api-key";

const response = await fetch(`%%URL%%/${plateID}/keys/mykey`, {
  headers: {
    "Authorization": apiKey,
    "Content-Type": "application/json"
  }
});

const data = await response.json();
console.log(data);
```

### cURL

```bash
curl -X GET "%%URL%%/my-plate/keys/mykey" \
  -H "Authorization: YOUR_API_KEY"
```

## Error Responses

Missing or invalid authorization returns `401 Unauthorized`:

```json
{
  "error": "missing_authorization",
  "message": "authorization header is required"
}
```

```json
{
  "error": "invalid_authorization",
  "message": "authorization denied"
}
```

## Security Notes

- Keep your API keys secure and never expose them in client-side code
- Rotate keys periodically
- Use HTTPS in production environments
- Implement key validation at the application level