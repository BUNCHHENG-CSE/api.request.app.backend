# FlowAPI Backend - REST API Documentation

## Base URL
```
http://localhost:8080/api
```

## Authentication
All protected endpoints require a Bearer token in the Authorization header:
```
Authorization: Bearer <your-jwt-token>
```

---

## Workspaces

### Create Workspace
```http
POST /workspaces
Authorization: Bearer <token>
Content-Type: application/json

{
  "name": "My Workspace",
  "description": "Development workspace",
  "icon": "🚀"
}
```

### List User Workspaces
```http
GET /workspaces
Authorization: Bearer <token>
```

### Get Workspace
```http
GET /workspaces/:id
Authorization: Bearer <token>
```

### Update Workspace
```http
PUT /workspaces/:id
Authorization: Bearer <token>
Content-Type: application/json

{
  "name": "Updated Name",
  "description": "Updated description"
}
```

### Delete Workspace
```http
DELETE /workspaces/:id
Authorization: Bearer <token>
```

---

## Workspace Members

### Add Member
```http
POST /workspaces/:id/members
Authorization: Bearer <token>
Content-Type: application/json

{
  "user_id": "user-uuid",
  "role": "editor"  // owner, editor, viewer
}
```

### Get Members
```http
GET /workspaces/:id/members
Authorization: Bearer <token>
```

### Update Member Role
```http
PUT /workspaces/:id/members/:userId/role
Authorization: Bearer <token>
Content-Type: application/json

{
  "role": "viewer"  // owner, editor, viewer
}
```

### Remove Member
```http
DELETE /workspaces/:id/members/:userId
Authorization: Bearer <token>
```

---

## Requests

### Create Request
```http
POST /requests
Authorization: Bearer <token>
Content-Type: application/json

{
  "collection_id": "collection-uuid",
  "name": "Get Users",
  "description": "Fetch all users",
  "method": "GET",
  "url": "https://api.example.com/users",
  "headers": [
    {
      "key": "Authorization",
      "value": "Bearer token",
      "enabled": true
    }
  ],
  "params": [
    {
      "key": "limit",
      "value": "10",
      "enabled": true
    }
  ],
  "body": "",
  "body_type": "none",
  "auth": {
    "type": "bearer",
    "bearer_token": "token-value"
  },
  "scripts": {
    "pre_request": "// Pre-request code",
    "post_response": "// Post-response code"
  },
  "settings": {
    "http_version": "auto",
    "strict_ssl": true,
    "follow_redirects": true,
    "max_redirects": 10
  }
}
```

### Get Request
```http
GET /requests/:id
Authorization: Bearer <token>
```

### Update Request
```http
PUT /requests/:id
Authorization: Bearer <token>
Content-Type: application/json

{
  "name": "Updated Name",
  "url": "https://new-url.com",
  "method": "POST",
  // ... other fields to update
}
```

### Delete Request
```http
DELETE /requests/:id
Authorization: Bearer <token>
```

### Execute Request
```http
POST /requests/:id/execute
Authorization: Bearer <token>
Content-Type: application/json

{
  "environment_id": "env-uuid",  // optional
  "override": {                   // optional
    "url": "https://override-url.com"
  }
}
```

**Response:**
```json
{
  "status_code": 200,
  "status_text": "200 OK",
  "headers": [
    {
      "key": "Content-Type",
      "value": "application/json"
    }
  ],
  "body": "{...}",
  "size": 1024,
  "time": 250,
  "cookies": []
}
```

### Duplicate Request
```http
POST /requests/:id/duplicate
Authorization: Bearer <token>
```

### Search Requests
```http
GET /workspaces/:workspaceId/requests/search?q=search-query
Authorization: Bearer <token>
```

### Get Request History
```http
GET /requests/:id/history?limit=20
Authorization: Bearer <token>
```

---

## Environments

### Create Environment
```http
POST /workspaces/:workspaceId/environments
Authorization: Bearer <token>
Content-Type: application/json

{
  "name": "Development",
  "active": true,
  "variables": [
    {
      "name": "API_KEY",
      "value": "secret-key",
      "secret": true
    },
    {
      "name": "BASE_URL",
      "value": "https://dev.example.com",
      "secret": false
    }
  ]
}
```

### List Environments
```http
GET /workspaces/:workspaceId/environments
Authorization: Bearer <token>
```

### Get Environment
```http
GET /environments/:id
Authorization: Bearer <token>
```

### Update Environment
```http
PUT /environments/:id
Authorization: Bearer <token>
Content-Type: application/json

{
  "name": "Staging",
  "variables": [...]
}
```

### Delete Environment
```http
DELETE /environments/:id
Authorization: Bearer <token>
```

### Set Active Environment
```http
POST /workspaces/:workspaceId/environments/:id/activate
Authorization: Bearer <token>
```

### Get Active Environment
```http
GET /workspaces/:workspaceId/environments/active
Authorization: Bearer <token>
```

---

## Flows

### Create Flow
```http
POST /workspaces/:workspaceId/flows
Authorization: Bearer <token>
Content-Type: application/json

{
  "name": "API Flow",
  "description": "Multi-step API workflow",
  "nodes": [
    {
      "id": "node-1",
      "type": "request",
      "request_id": "request-uuid",
      "position": {"x": 0, "y": 0},
      "data": {}
    }
  ],
  "edges": [
    {
      "id": "edge-1",
      "source": "node-1",
      "target": "node-2"
    }
  ]
}
```

### Get Flow
```http
GET /flows/:id
Authorization: Bearer <token>
```

### List Flows
```http
GET /workspaces/:workspaceId/flows
Authorization: Bearer <token>
```

### Update Flow
```http
PUT /flows/:id
Authorization: Bearer <token>
Content-Type: application/json

{
  "name": "Updated Flow",
  "nodes": [...],
  "edges": [...]
}
```

### Delete Flow
```http
DELETE /flows/:id
Authorization: Bearer <token>
```

### Execute Flow
```http
POST /flows/:id/execute
Authorization: Bearer <token>
Content-Type: application/json

{
  "context_var": "value"  // optional execution context
}
```

**Response:**
```json
{
  "results": {
    "node-1": {
      "status_code": 200,
      "body": "..."
    }
  }
}
```

---

## Error Response Format

All error responses follow this format:

```json
{
  "error": "ERROR_CODE",
  "message": "Human readable error message",
  "code": 400
}
```

### Common Error Codes
- `400` - Bad Request (validation error)
- `401` - Unauthorized (missing/invalid token)
- `403` - Forbidden (insufficient permissions)
- `404` - Not Found
- `500` - Internal Server Error

---

## Request Execution Details

### Authorization Types

#### Bearer Token
```json
{
  "type": "bearer",
  "bearer_token": "your-token"
}
```

#### Basic Auth
```json
{
  "type": "basic",
  "basic_username": "username",
  "basic_password": "password"
}
```

#### API Key
```json
{
  "type": "api_key",
  "api_key_key": "X-API-Key",
  "api_key_value": "your-key",
  "api_key_in": "header"  // header or query
}
```

#### OAuth2
```json
{
  "type": "oauth2",
  "oauth2": {
    "grant_type": "authorization_code",
    "auth_url": "https://...",
    "token_url": "https://...",
    "client_id": "...",
    "client_secret": "...",
    "scope": "read write",
    "redirect_uri": "http://localhost:3000/callback"
  }
}
```

### Body Types
- `none` - No body
- `json` - JSON body
- `form` - Form body (application/x-www-form-urlencoded)
- `form-data` - Multipart form data
- `raw` - Raw text
- `binary` - Binary data
- `text` - Plain text
- `xml` - XML
- `graphql` - GraphQL query

---

## Pagination

For list endpoints, use query parameters:
```
?page=1&limit=20
```

Response includes pagination info:
```json
{
  "data": [...],
  "total": 100,
  "page": 1,
  "limit": 20,
  "total_page": 5
}
```

---

## Environment Variable Substitution

Use `{{variable_name}}` syntax in requests:

```
URL: https://api.example.com/{{version}}/users
Headers: Authorization: Bearer {{API_TOKEN}}
Body: {"key": "{{SECRET_KEY}}"}
```

Variables are replaced from the active environment at execution time.
