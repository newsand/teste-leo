# Users API (Go)

A RESTful JSON HTTP API for user management, built with Go, Chi router, and SQLite.

## Features

- Full CRUD operations for users
- Paginated list endpoint with metadata
- Password hashing with bcrypt
- SQLite database with proper indexes
- Input validation
- JSON error responses

## Requirements

- Go 1.22+

## Build & Run

```bash
# Navigate to the Go directory
cd go

# Download dependencies
go mod tidy

# Build the application
go build -o users-api ./cmd/api

# Run the server (defaults to port 8080)
./users-api

# Or run directly with go run
go run ./cmd/api
```

### Configuration

| Flag/Env | Default | Description |
|----------|---------|-------------|
| `-port` / `PORT` | `8080` | Server port |
| `-db` / `DB_PATH` | `users.db` | SQLite database path |

## API Endpoints

### Create User
```bash
curl -X POST http://localhost:8080/users \
  -H "Content-Type: application/json" \
  -d '{"name": "John Doe", "email": "john@example.com", "password": "secret123"}'
```

Response (201 Created):
```json
{
  "id": 1,
  "name": "John Doe",
  "email": "john@example.com",
  "created_at": "2024-01-15T10:30:00Z",
  "updated_at": "2024-01-15T10:30:00Z"
}
```

### List Users (Paginated)
```bash
curl "http://localhost:8080/users?page=1&limit=10"
```

Response (200 OK):
```json
{
  "data": [
    {
      "id": 1,
      "name": "John Doe",
      "email": "john@example.com",
      "created_at": "2024-01-15T10:30:00Z",
      "updated_at": "2024-01-15T10:30:00Z"
    }
  ],
  "pagination": {
    "page": 1,
    "limit": 10,
    "total": 1,
    "total_pages": 1
  }
}
```

### Get User by ID
```bash
curl http://localhost:8080/users/1
```

Response (200 OK):
```json
{
  "id": 1,
  "name": "John Doe",
  "email": "john@example.com",
  "created_at": "2024-01-15T10:30:00Z",
  "updated_at": "2024-01-15T10:30:00Z"
}
```

### Update User (PUT - full replace)
```bash
curl -X PUT http://localhost:8080/users/1 \
  -H "Content-Type: application/json" \
  -d '{"name": "Jane Doe", "email": "jane@example.com"}'
```

### Update User (PATCH - partial update)
```bash
curl -X PATCH http://localhost:8080/users/1 \
  -H "Content-Type: application/json" \
  -d '{"name": "Jane Updated"}'
```

### Delete User
```bash
curl -X DELETE http://localhost:8080/users/1
```

Response (200 OK):
```json
{
  "message": "User deleted successfully"
}
```

## Error Responses

All errors return JSON with an `error` field:

```json
{
  "error": "User not found"
}
```

| Status Code | Description |
|-------------|-------------|
| 200 | Success |
| 201 | Created |
| 400 | Bad Request (validation error, duplicate email) |
| 404 | Not Found |
| 500 | Internal Server Error |

## Database Schema

```sql
CREATE TABLE users (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    email TEXT NOT NULL UNIQUE,
    password TEXT NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE UNIQUE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_created_at ON users(created_at DESC);
```

## Running Tests

```bash
cd go
go test ./... -v
```

## Project Structure

```
go/
├── cmd/
│   └── api/
│       └── main.go          # Application entry point
├── internal/
│   ├── database/
│   │   ├── sqlite.go        # Database initialization
│   │   └── users.go         # User repository
│   ├── handlers/
│   │   ├── users.go         # HTTP handlers
│   │   └── users_test.go    # Handler tests
│   └── models/
│       └── user.go          # Data models
├── go.mod
├── go.sum
└── README.md
```
