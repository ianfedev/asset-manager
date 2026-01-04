# Middleware Documentation

## Overview
The Asset Manager uses Fiber middleware to enforce security and improve observability.

## Authentication (API Key)
All API endpoints are protected by an API Key.
- **Header**: `X-API-Key`
- **Configuration**: `SERVER_API_KEY` in `.env`.
- **Behavior**:
    - If the header is missing or incorrect, the server returns `401 Unauthorized`.

## Ray ID (Request Tracing)
Every request is assigned a unique identifier (Ray ID) for tracing purposes.
- **Algorithm**: UUID v4.
- **Header**: The server includes `X-Ray-ID` in every response.
- **Context**: The Ray ID is stored in the Fiber context locals under the key `ray_id`.
- **Logging**: The logger automatically includes the Ray ID in all log entries associated with a request if using the request-scoped logger.

## Usage

### Client Request
```bash
curl -H "X-API-Key: your-secret-key" http://localhost:8080/some/path
```

### Server Response
```http
HTTP/1.1 200 OK
X-Ray-ID: 550e8400-e29b-41d4-a716-446655440000
...
```
