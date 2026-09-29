# Users API

RESTful JSON HTTP API for user management with SQLite storage.

## Features

- CRUD operations for users
- Password hashing with bcrypt
- Paginated listing
- Input validation
- SQLite database with optimized indexes

## Requirements

- Node.js 18+
- npm

## Installation

```bash
npm install
```

## Running the Server

```bash
# Production
npm start

# Development (with auto-reload)
npm run dev
```

The server starts on port 3000 by default. Set `PORT` environment variable to change it.

## API Endpoints

### Create User
```bash
curl -X POST http://localhost:3000/users \
  -H "Content-Type: application/json" \
  -d '{"name": "John Doe", "email": "john@example.com", "password": "secret123"}'
```

Response (201 Created):
```json
{
  "id": 1,
  "name": "John Doe",
  "email": "john@example.com",
  "created_at": "2024-01-01 12:00:00",
  "updated_at": "2024-01-01 12:00:00"
}
```

### List Users (Paginated)
```bash
curl "http://localhost:3000/users?page=1&limit=10"
```

Response (200 OK):
```json
{
  "data": [
    {
      "id": 1,
      "name": "John Doe",
      "email": "john@example.com",
      "created_at": "2024-01-01 12:00:00",
      "updated_at": "2024-01-01 12:00:00"
    }
  ],
  "pagination": {
    "page": 1,
    "limit": 10,
    "total": 1,
    "totalPages": 1,
    "hasNextPage": false,
    "hasPreviousPage": false
  }
}
```

### Get User by ID
```bash
curl http://localhost:3000/users/1
```

Response (200 OK):
```json
{
  "id": 1,
  "name": "John Doe",
  "email": "john@example.com",
  "created_at": "2024-01-01 12:00:00",
  "updated_at": "2024-01-01 12:00:00"
}
```

### Update User (PUT - full update)
```bash
curl -X PUT http://localhost:3000/users/1 \
  -H "Content-Type: application/json" \
  -d '{"name": "John Updated", "email": "john.new@example.com"}'
```

### Update User (PATCH - partial update)
```bash
curl -X PATCH http://localhost:3000/users/1 \
  -H "Content-Type: application/json" \
  -d '{"name": "John Patched"}'
```

### Delete User
```bash
curl -X DELETE http://localhost:3000/users/1
```

Response (200 OK):
```json
{
  "message": "User deleted successfully"
}
```

## Error Responses

### 400 Bad Request (Validation Error)
```json
{
  "error": "Validation failed",
  "details": ["email must be a valid email address"]
}
```

### 404 Not Found
```json
{
  "error": "User not found"
}
```

### 500 Internal Server Error
```json
{
  "error": "Internal server error"
}
```

## Testing

### Run unit tests
```bash
npm test
```

### Run smoke tests (requires running server)
```bash
# In one terminal:
npm start

# In another terminal:
chmod +x scripts/smoke-test.sh
./scripts/smoke-test.sh
```

## Database

SQLite database is stored at `data/users.db`. The schema:

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

## Security Notes

- Passwords are hashed using bcrypt (10 rounds)
- Password hashes are never returned in API responses
- Email uniqueness is enforced at the database level
