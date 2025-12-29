# Auth Service

A lightweight authentication service built with Go standard library.

## Setup

1. **Install dependencies:**
   ```bash
   go mod download
   ```

2. **Configure environment variables:**
   ```bash
   export DB_HOST=localhost
   export DB_PORT=5432
   export DB_USER=auth_user
   export DB_PASSWORD=your_password
   export DB_NAME=auth_db
   export DB_SSLMODE=disable

   export SMTP_HOST=smtp.example.com
   export SMTP_PORT=587
   export SMTP_USER=your_email@example.com
   export SMTP_PASSWORD=your_smtp_password
   export SMTP_FROM=noreply@example.com

   export SERVER_PORT=8080
   export SERVER_HOST=0.0.0.0
   ```

3. **Initialize database:**
   ```bash
   psql -d auth_db -f schema.sql
   ```

## Build and Run

```bash
go build ./cmd/api
./api
```

Or run directly:
```bash
go run ./cmd/api
```

## API Endpoints

### Authentication
- `GET /auth` - Render authentication page
- `POST /auth/email` - Submit email for login/signup
- `POST /auth/password` - Submit password for login/signup
- `GET /auth/confirm` - Confirm email address

### Password Reset
- `GET /reset` - Render password reset page
- `POST /reset/request` - Request password reset
- `GET /reset/confirm` - Render confirmation page
- `POST /reset/complete` - Complete password reset

### Account Management
- `GET /account` - Render account page (authenticated)
- `POST /account/email` - Change email (authenticated)
- `POST /account/password` - Change password (authenticated)
- `POST /account/delete` - Delete account (authenticated)

## Development

### Format code
```bash
go fmt ./...
```

### Run tests
```bash
go test ./...
```

### Run tests with coverage
```bash
go test -cover ./...
```

## License

MIT
