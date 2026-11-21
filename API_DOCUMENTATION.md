# FlowAPI HTTP API

Frontend integration contract for the Go backend.

## Base URLs

- API: `http://localhost:8080/api/v1`
- Health: `http://localhost:8080/health`

`SERVER_PORT` controls the port and defaults to `8080`.

## Common response envelope

Success:

```json
{
  "success": true,
  "message": "Operation completed successfully",
  "data": {}
}
```

Error:

```json
{
  "success": false,
  "error": "Human-readable error"
}
```

For successful delete and logout operations, `data` is an empty object.

## Authentication

Only registration, login, and the health check are public. All other routes require:

```http
Authorization: Bearer <jwt-token>
Accept: application/json
```

JWTs expire 24 hours after issuance. `POST /users/refresh` returns a new token. Logout is stateless: the frontend must delete its token/cookie; the server does not maintain a revocation list.

## Endpoint matrix

All paths below except `/health` are relative to `/api/v1`.

### System

| Method | Path | Purpose | Auth |
|---|---|---|---|
| `GET` | `/health` | Liveness check | No |
| `GET` | `/search?q=users&workspace_id=:id&limit=20` | Search owned flows, requests, collections, and environments | Yes |

### Authentication and users

| Method | Path | Purpose | Auth |
|---|---|---|---|
| `POST` | `/users/register` | Register an account | No |
| `POST` | `/users/login` | Get a JWT and user | No |
| `POST` | `/users/refresh` | Replace the current JWT | Yes |
| `POST` | `/users/logout` | Confirm client-side logout | Yes |
| `POST` | `/users/change-password` | Verify and replace the authenticated user's password | Yes |
| `GET` | `/users/:id` | Get the authenticated user's profile | Yes |
| `PATCH` | `/users/:id` | Update the authenticated user's username | Yes |

### Workspaces

| Method | Path | Purpose |
|---|---|---|
| `GET` | `/workspaces` | List workspaces owned by the user |
| `POST` | `/workspaces` | Create a workspace |
| `GET` | `/workspaces/:id` | Get an owned workspace |
| `PATCH` | `/workspaces/:id` | Rename an owned workspace |
| `DELETE` | `/workspaces/:id` | Delete workspace and its child resources |
| `GET` | `/workspaces/:id/flows` | List workspace flows |
| `GET` | `/workspaces/:id/environments` | List workspace environments |
| `GET` | `/workspaces/:id/collections` | List workspace collections |

### Flows

| Method | Path | Purpose |
|---|---|---|
| `POST` | `/flows` | Create a flow |
| `GET` | `/flows/:id` | Get a flow |
| `PATCH` | `/flows/:id` | Update flow metadata or graph data |
| `DELETE` | `/flows/:id` | Delete a flow and its requests |
| `GET` | `/flows/:id/requests` | List stored requests in a flow |
| `POST` | `/flows/:id/run` | Execute all requests in a flow sequentially |

### Stored requests

| Method | Path | Purpose |
|---|---|---|
| `POST` | `/requests` | Create a stored request |
| `GET` | `/requests/:id` | Get a stored request |
| `PATCH` | `/requests/:id` | Update a stored request |
| `DELETE` | `/requests/:id` | Permanently delete a stored request |
| `POST` | `/requests/:id/execute` | Execute a stored request from the backend |
| `POST` | `/requests/:id/duplicate` | Duplicate a stored request |
| `POST` | `/requests/bulk-delete` | Delete up to 100 requests |

### Environments

| Method | Path | Purpose |
|---|---|---|
| `POST` | `/environments` | Create an environment |
| `GET` | `/environments/:id` | Get an environment |
| `PATCH` | `/environments/:id` | Update name or variables |
| `DELETE` | `/environments/:id` | Delete an environment |

### Collections

| Method | Path | Purpose |
|---|---|---|
| `POST` | `/collections` | Create a collection |
| `GET` | `/collections/:id` | Get a collection |
| `PATCH` | `/collections/:id` | Update a collection |
| `DELETE` | `/collections/:id` | Delete a collection |
| `POST` | `/collections/:id/requests/:requestId` | Associate a request with a collection |
| `DELETE` | `/collections/:id/requests/:requestId` | Remove a request from a collection |

### Local Bruno-style storage

| Method | Path | Purpose |
|---|---|---|
| `POST` | `/workspaces/:id/local/export` | Export/synchronize a database workspace to local Bruno files |
| `GET` | `/local-collections` | List local collections owned by the authenticated user |
| `GET` | `/local-collections/:workspaceId` | Get a local collection's path and sync metadata |
| `POST` | `/local-collections/:workspaceId/import` | Import local `.bru` edits into PostgreSQL |
| `DELETE` | `/local-collections/:workspaceId` | Delete only the local filesystem copy |

### Execution history

| Method | Path | Purpose |
|---|---|---|
| `GET` | `/history?page=1&limit=20&workspace_id=:id` | List execution history; workspace filter is optional |
| `GET` | `/history/:id` | Get a complete history entry |
| `DELETE` | `/history` | Clear the authenticated user's history |

## Authentication payloads

### Register

```http
POST /api/v1/users/register
Content-Type: application/json
```

```json
{
  "username": "alice",
  "email": "alice@example.com",
  "password": "at-least-6-characters"
}
```

### Login

```http
POST /api/v1/users/login
Content-Type: application/json
```

```json
{
  "email": "alice@example.com",
  "password": "at-least-6-characters"
}
```

Login response data:

```json
{
  "token": "jwt-token",
  "user": {
    "id": "user-uuid",
    "username": "alice",
    "email": "alice@example.com",
    "created_at": "2026-10-01T00:00:00Z",
    "updated_at": "2026-10-01T00:00:00Z"
  }
}
```

### Refresh

```http
POST /api/v1/users/refresh
Authorization: Bearer <current-token>
```

Response data:

```json
{
  "token": "new-jwt-token"
}
```

### Update profile

```http
PATCH /api/v1/users/:id
Content-Type: application/json
```

```json
{
  "username": "new-name"
}
```

### Change password

```http
POST /api/v1/users/change-password
Authorization: Bearer <token>
Content-Type: application/json
```

```json
{
  "current_password": "current-password",
  "new_password": "new-secure-password",
  "confirm_password": "new-secure-password"
}
```

Rules:

- All three fields are required.
- The current password must be correct.
- The new password must contain at least 8 characters.
- `new_password` and `confirm_password` must match.
- The new password must be different from the current password.
- Passwords are stored as bcrypt hashes and are never included in API responses.

Successful response:

```json
{
  "success": true,
  "message": "Password changed successfully",
  "data": {}
}
```

Existing JWTs remain valid until their normal 24-hour expiration. The frontend should log the user out after a successful password change if it requires all local sessions to reauthenticate.

## Workspace payloads

Create and rename use the same shape. `owner_id` is always taken from the JWT and must not be supplied by the frontend.

```json
{
  "name": "My Workspace"
}
```

Deleting a workspace also deletes its flows, stored requests, environments, and collections.

## Flow payloads

Create:

```json
{
  "workspace_id": "workspace-uuid",
  "name": "Checkout flow",
  "description": "Optional description",
  "nodes": "[]",
  "edges": "[]"
}
```

`nodes` and `edges` are JSON-encoded strings. They default to empty strings when omitted.

Patch accepts any subset:

```json
{
  "name": "Updated flow",
  "description": "Updated description",
  "nodes": "[{\"id\":\"node-1\"}]",
  "edges": "[]"
}
```

## Stored-request payloads

Create:

```json
{
  "flow_id": "flow-uuid",
  "collection_id": "collection-uuid",
  "name": "List users",
  "position": 0,
  "method": "GET",
  "url": "https://api.example.com/users",
  "headers": "{\"Accept\":\"application/json\"}",
  "query_params": "{\"limit\":\"20\"}",
  "body": "",
  "body_type": "none",
  "auth": "{}",
  "scripts": "{}",
  "settings": "{}"
}
```

Only `flow_id`, `method`, and `url` are required. `collection_id` is optional. The collection and flow must belong to the same workspace.

`PATCH /requests/:id` accepts any subset of:

- `flow_id`
- `name`
- `position`
- `method`
- `url`
- `headers`
- `query_params`
- `body`
- `body_type`
- `auth`
- `scripts`
- `settings`

The JSON-like fields remain JSON-encoded strings for compatibility with the current frontend types. Both JSON objects (e.g., `{"key": "value"}`) and arrays of key-value items (e.g., `[{"key": "k", "value": "v", "enabled": true}]`) with string, numeric, or boolean values are supported.

`auth` supports Bearer tokens (`{"type":"bearer","token":"..."}`), Basic authentication (`{"type":"basic","username":"...","password":"..."}`), and API Keys (`{"type":"apikey","key":"...","value":"...","in":"header|query"}`). When present, auth credentials are automatically applied to outbound executions unless explicitly overridden in `headers`. `scripts` and `settings` remain persisted metadata in this version.

### Execute a stored request

```http
POST /api/v1/requests/:id/execute
Authorization: Bearer <token>
```

The body is optional. Pass an environment to resolve `{{variable}}` placeholders in the URL, headers, query parameters, and body:

```json
{
  "environment_id": "environment-uuid"
}
```

Response data:

```json
{
  "history_id": "history-uuid",
  "status": 200,
  "statusText": "200 OK",
  "time": 142,
  "size": "1240B",
  "headers": {
    "Content-Type": "application/json"
  },
  "body": "{\"ok\":true}",
  "truncated": false
}
```

Execution behavior:

- Only absolute `http://` and `https://` URLs are accepted.
- Timeout is 30 seconds.
- Response bodies are limited to 2 MB and marked with `truncated: true` when capped.
- Stored headers must be a JSON object whose values are strings.
- Stored `query_params` are merged into the URL and must be a JSON object whose values are strings.
- GET and HEAD requests do not send the stored body.
- Redirects use Go's standard client behavior.
- Every success or failure creates a persistent history entry.
- Private, loopback, link-local, and multicast destinations are blocked unless `ALLOW_PRIVATE_REQUESTS=true`.

### Run a flow

```http
POST /api/v1/flows/:id/run
```

```json
{
  "environment_id": "optional-environment-uuid",
  "stop_on_error": true
}
```

Requests run sequentially in their stored order. Each execution creates a history entry. `stop_on_error` defaults to `true`.

Use the request `position` field to control flow execution order. Lower positions run first; equal positions use creation time.

### Bulk delete requests

```json
{
  "ids": ["request-uuid-1", "request-uuid-2"]
}
```

## Environment payloads

Create:

```json
{
  "workspace_id": "workspace-uuid",
  "name": "Development",
  "variables": "{\"BASE_URL\":\"https://dev.example.com\"}"
}
```

Patch accepts `name` and/or `variables`:

```json
{
  "variables": "{\"BASE_URL\":\"https://staging.example.com\"}"
}
```

`variables` is currently a JSON-encoded object of string values. Selecting an environment during request or flow execution resolves `{{VARIABLE_NAME}}` placeholders.

## Collection payloads

Create:

```json
{
  "workspace_id": "workspace-uuid",
  "name": "User APIs",
  "description": "Optional description"
}
```

Patch accepts `name` and/or `description`.

Associating a request sets its `collection_id`. Removing it sets `collection_id` to `null`. Deleting a collection does not delete its requests; their `collection_id` is cleared.

## Local Bruno-style collections

Local storage is hybrid: PostgreSQL remains authoritative for users, ownership, and execution history, while a workspace can be synchronized to human-readable files for local editing and Git workflows.

### Export a workspace

```http
POST /api/v1/workspaces/:id/local/export
Authorization: Bearer <token>
```

No body is required. The response includes the absolute backend-local path:

```json
{
  "workspace_id": "workspace-uuid",
  "name": "My Workspace",
  "path": "/data/collections/workspace-uuid",
  "updated_at": "2026-10-01T12:00:00Z"
}
```

Generated layout:

```text
workspace-uuid/
├── bruno.json
├── flowapi.json
├── environments/
│   └── environment-uuid-development.bru
└── requests/
    └── flow-uuid/
        └── 0001-list-users-requestid.bru
```

- `bruno.json` makes the directory recognizable as a Bruno collection.
- Request and environment files use Bruno's readable `.bru` structure.
- `flowapi.json` preserves UUIDs, flow graphs, collection associations, and fields that do not map directly to Bru syntax.
- Re-exporting updates generated files but does not remove unrelated user-created files.

### Import local changes

```http
POST /api/v1/local-collections/:workspaceId/import
Authorization: Bearer <token>
```

Import reads changes to request names, order, methods, URLs, headers, query parameters, bodies, and environment variables from `.bru` files. It then creates or updates the corresponding database records.

Import response data:

```json
{
  "flows": 2,
  "requests": 12,
  "environments": 2,
  "collections": 3
}
```

Import is non-destructive: removing a `.bru` file does not automatically delete the database request. Use the API delete endpoint for explicit deletion, then export again.

### Export Bruno collection as ZIP archive

```http
GET /api/v1/local-collections/:workspaceId/export-zip
Authorization: Bearer <token>
```

Exports the entire Bruno collection directory into a downloadable `.zip` archive (`application/zip`) formatted with `bruno.json`, `environments/*.bru`, and `requests/*/*.bru`.

### Import Bruno collection from ZIP archive

```http
POST /api/v1/local-collections/:workspaceId/import-zip
Authorization: Bearer <token>
Content-Type: multipart/form-data
```

Upload a `.zip` archive containing a Bruno collection. Safely unpacks the files into the workspace directory on local disk and updates all database records.

### Browse and edit local Bruno collection files

- **List collection file tree:** `GET /api/v1/local-collections/:workspaceId/files`
  Returns an array of `{ "path": "...", "is_dir": bool, "size": int, "modified": time }`.
- **Read local file:** `GET /api/v1/local-collections/:workspaceId/file?path=bruno.json`
- **Write local file:** `PUT /api/v1/local-collections/:workspaceId/file` with body `{"path": "requests/.../request.bru", "content": "..."}`.

### Import from cURL

```http
POST /api/v1/requests/import-curl
Authorization: Bearer <token>
Content-Type: application/json

{
  "flow_id": "flow-uuid",
  "curl": "curl -X POST 'https://api.example.com/users' -H 'Authorization: Bearer token' -d '{\"name\":\"Alice\"}'"
}
```

Parses method, URL, headers, query params, body, and auth into a stored request.

### Import from Postman Collection (v2.1)

```http
POST /api/v1/collections/import-postman
Authorization: Bearer <token>
Content-Type: application/json (or multipart/form-data with "file")

{
  "workspace_id": "workspace-uuid",
  "collection": { ...postman v2.1 collection JSON... }
}
```

Imports all folders, requests, headers, bodies, and auth credentials from a Postman v2.1 collection into the workspace.

### Generate Code Snippets

```http
GET /api/v1/requests/:id/snippet?language=curl
Authorization: Bearer <token>
```

Supported languages: `curl`, `javascript` (fetch), `python` (requests), `golang` (net/http). Returns an executable code snippet.

### Mock Server Engine

```http
ANY /api/v1/mock/:workspaceId/*path
```

Public or authenticated mock endpoint. Dynamically matches simulated routes against stored requests in the workspace. If the request defines mock settings in its `settings` field (`mock_status`, `mock_body`, `mock_headers`, `mock_content_type`), it returns that response. Otherwise, it returns an intelligent simulated response.

### Declarative Assertions & Test Runner

Requests support declarative test assertions configured in `scripts` or `settings` as a JSON array:

```json
[
  {"target": "status", "operator": "equals", "expected": 200},
  {"target": "time", "operator": "less_than", "expected": 1000},
  {"target": "header", "property": "Content-Type", "operator": "contains", "expected": "application/json"},
  {"target": "body", "operator": "contains", "expected": "success"},
  {"target": "json_field", "property": "data.id", "operator": "exists"},
  {"target": "json_field", "property": "status", "operator": "equals", "expected": "ok"}
]
```

When executing a request or running a flow, the engine evaluates each assertion against the response and returns `assertions`, `tests_passed`, and `tests_failed` in the result payload.

### Delete the local copy

`DELETE /local-collections/:workspaceId` removes only the directory beneath `LOCAL_COLLECTIONS_ROOT`. It does not delete the PostgreSQL workspace or its resources.

### Filesystem security

- Collection directories are always named using workspace UUIDs.
- All paths are resolved beneath `LOCAL_COLLECTIONS_ROOT`; absolute paths and traversal outside the root are rejected.
- Imports verify workspace ownership and reject resource IDs belonging to another workspace.
- File writes use temporary files and rename operations to reduce partial-write corruption.
- Environment variables are exported as readable text. Do not commit secrets; the default project `.gitignore` excludes `data/collections`.

## Authorization and ownership

- Users can read or update only their own profile.
- Workspaces are scoped to the JWT user.
- Flow, request, environment, and collection operations validate ownership through the parent workspace.
- A request cannot be moved into a flow or collection belonging to another workspace or user.
- Access failures are returned as `404` to avoid exposing another user's resource IDs.

## Runtime configuration

| Variable | Default | Purpose |
|---|---|---|
| `SERVER_PORT` | `8080` | API listening port |
| `DATABASE_URL` | Local PostgreSQL DSN | PostgreSQL connection |
| `JWT_SECRET` | Development fallback | JWT signing secret; always replace in production |
| `ALLOWED_ORIGINS` | `http://localhost:3000` | Comma-separated browser origins allowed by CORS |
| `ALLOW_PRIVATE_REQUESTS` | `false` | Permit request execution against localhost/private networks |
| `LOCAL_COLLECTIONS_ROOT` | `./data/collections` | Root directory for Bruno-style local collections |

The server also applies request IDs, security headers, HTTP timeouts, database connection pooling, controlled proxy trust, and graceful shutdown.

## Status codes

| Code | Meaning |
|---|---|
| `200` | Successful read, update, delete, refresh, logout, association, or execution |
| `201` | Resource created |
| `400` | Invalid UUID, JSON, URL, method, or request validation error |
| `401` | Missing, expired, or invalid bearer token / credentials |
| `403` | Authenticated user attempted a forbidden operation or ownership mismatch |
| `404` | Resource not found or not owned by the authenticated user |
| `409` | Conflict (e.g., email already in use) |
| `502` | Outbound stored-request execution failed (network/target unreachable) |

## Recommended frontend sequence

1. Register or login.
2. Store `data.token` in an HTTP-only cookie.
3. Send the bearer token to every protected endpoint.
4. Create/select a workspace.
5. Load its flows, environments, and collections in parallel.
6. Load requests for each flow.
7. Refresh the token before its 24-hour expiry and replace the cookie.
8. Delete the local cookie during logout.

## Still not implemented

The following items from `API_GAPS.md` need additional schemas or product decisions and are not available yet:

- OAuth providers, account linking, and server-side JWT revocation.
- Workspace members, invitations, roles, presence, and activity.
- Nested folders and manual ordering.
- Encrypted environment secrets and multi-scope variable precedence.
- Persisted flow-run summaries, script/test execution, monitors, schedules, and performance testing.
- Pagination/filter/sort metadata on resource list endpoints beyond history search limits.
- OpenAPI, Postman, and cURL import/export or code generation beyond the implemented Bruno-style local format.
- Organization-specific outbound allowlists and audit policy beyond the built-in private-network SSRF guard.

Do not integrate against unlisted routes.
