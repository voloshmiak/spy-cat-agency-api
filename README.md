# Spy Cat Agency API

A REST API for managing spy cats, their missions, and targets. Built with Go, PostgreSQL, and Docker.

## Prerequisites

Before you begin, ensure you have the following installed on your system:

- **Go** (version 1.19 or later)
- **Docker** and **Docker Compose**

## Quick Start

You can get the entire application up and running with just 2 commands:

```bash
# 1. Clone and navigate to the project
git clone https://github.com/voloshmiak/spy-cat-agency-api.git && cd spy-cat-agency-api

# 2. Start everything with Docker Compose
docker-compose up --build
```

The API will be available at `http://localhost:8080/api/v1`

## API Endpoints

Once the application is running, you can test the following endpoints:

### Cats
- `GET /api/v1/cats` - List all spy cats
- `POST /api/v1/cats` - Create a new spy cat
- `GET /api/v1/cats/{id}` - Get a specific cat
- `PATCH /api/v1/cats/{id}` - Update cat's salary
- `DELETE /api/v1/cats/{id}` - Delete a cat

### Missions
- `GET /api/v1/missions` - List all missions
- `POST /api/v1/missions` - Create a new mission
- `GET /api/v1/missions/{id}` - Get a specific mission
- `PATCH /api/v1/missions/{id}` - Update a mission
- `DELETE /api/v1/missions/{id}` - Delete a mission

### Targets
- `POST /api/v1/missions/{missionId}/targets` - Add a target to a mission
- `PATCH /api/v1/missions/{missionId}/targets/{targetId}` - Update a target
- `DELETE /api/v1/missions/{missionId}/targets/{targetId}` - Delete a target

## Testing the API

### Create a Spy Cat
```bash
curl -X POST http://localhost:8080/api/v1/cats \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Agent Whiskers",
    "years_of_experience": 5,
    "breed": "Siamese",
    "salary": 75000.50
  }'
```

### Create a Mission
```bash
curl -X POST http://localhost:8080/api/v1/missions \
  -H "Content-Type: application/json" \
  -d '{
    "targets": [
      {
        "name": "Infiltrate Evil Corp",
        "country": "USA"
      }
    ]
  }'
```

### List All Cats
```bash
curl http://localhost:8080/api/v1/cats
```

## Database

The application uses PostgreSQL with automatic migrations. The database schema includes:

- **cats** - Spy cat information
- **missions** - Mission details
- **targets** - Mission targets

Database migrations are automatically applied when the application starts.

### Database Migrations

Migrations are located in the `migrations/` directory and are automatically applied on startup.

### Stopping the Application

```bash
# Stop Docker Compose services
docker-compose down

# Remove volumes (this will delete all data)
docker-compose down -v
```

## API Documentation

The API follows OpenAPI 3.0 specification. You can find the detailed API documentation in the `api/` directory.

## Project Structure

```
spy-cat-agency-api/
├── cmd/api/          # Application entry point
├── internal/         # Internal application code
│   ├── cat/         # Cat-related handlers, services, and models
│   ├── mission/     # Mission-related handlers, services, and models
│   ├── db/          # Database connection and utilities
│   └── middleware/  # HTTP middleware
├── migrations/       # Database migration files
├── config/          # Configuration files
├── api/             # API documentation
├── docker-compose.yaml
├── Dockerfile
└── README.md
```