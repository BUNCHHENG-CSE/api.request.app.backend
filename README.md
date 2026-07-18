# FlowAPI Backend - Go Implementation

A comprehensive REST API backend for FlowAPI, built with Go.

## Project Structure

```
api.request.app.backend/
├── cmd/
│   └── api/
│       └── main.go                 
├── internal/
│   ├── domain/                     
│   │   ├── collection.go
│   │   ├── environment.go
│   │   ├── flow.go
│   │   ├── request.go
│   │   ├── user.go
│   │   └── workspace.go
│   ├── application/                
│   │   ├── collection_service.go
│   │   ├── environment_service.go
│   │   ├── flow_service.go
│   │   ├── request_service.go
│   │   ├── user_service.go
│   │   └── workspace_service.go
│   ├── infrastructure/            
│   │   ├── database/
│   │   │   └── postgres.go         
│   │   ├── repositories/           
│   │   │   ├── collection_repo.go
│   │   │   ├── environment_repo.go
│   │   │   ├── flow_repo.go
│   │   │   ├── request_repo.go
│   │   │   ├── user_repo.go
│   │   │   └── workspace_repo.go
│   │   └── config/
│   │       └── config.go
│   └── interfaces/                 
│       ├── api/
│       │   ├── handlers/
│       │   │   ├── collection_handler.go
│       │   │   ├── environment_handler.go
│       │   │   ├── flow_handler.go
│       │   │   ├── request_handler.go
│       │   │   ├── requests.go
│       │   │   ├── user_handler.go
│       │   │   └── workspace_handler.go
│       │   ├── routes/
│       │   │   └── routes.go       
│       │   ├── middleware/
│       │   │   └── auth.go
│       │   └── response/
│       │       └── response.go
├── .env
├── docker-compose.yml
├── go.mod
└── go.sum
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
