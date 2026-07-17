# FlowAPI Backend - Go Implementation

A comprehensive REST API backend for FlowAPI, built with Go.

## Project Structure

```
backend/
├── cmd/
│   └── api/
│       └── main.go                 # Entry point
├── internal/
│   ├── models/                     # Data models
│   ├── entities/                   # Database entities
│   ├── services/                   # Business logic
│   ├── controllers/                # HTTP handlers
│   ├── repositories/               # Data access layer
│   ├── middleware/                 # HTTP middleware
│   ├── config/                     # Configuration
│   └── utils/                      # Utilities
├── migrations/                     # Database migrations
├── pkg/
│   └── errors/                     # Custom errors
├── go.mod
├── go.sum
└── docker-compose.yml
```

## Key Features

- User Authentication & Authorization
- Workspace & Collection Management
- API Request Management with full HTTP support
- Environment Variable Management
- Authorization Types (Bearer, Basic, API Key, OAuth2)
- Request/Response Scripts
- Team Collaboration
- Request History & Search
- Visual Flow Builder
- OpenAPI Spec Management

## Database

PostgreSQL with proper indexing and relationships.

## Deployment

Docker-ready with docker-compose for local development.
