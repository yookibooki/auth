# Agent Guidelines

This file contains guidelines for agentic coding agents working on this Go authentication service.

## Commands

### Build
```bash
go build ./cmd/main.go
```

### Run the API
```bash
go run ./cmd/main.go
```

### Format
```bash
go fmt ./...
```

### Database
```bash
# Apply schema to PostgreSQL
psql -d auth_db -f schema.sql
```

## Code Style Guidelines

### Go Standard Library Only
- No external dependencies unless absolutely necessary
- Use stdlib packages: `net/http`, `database/sql`, `html/template`, `context`, `crypto/bcrypt`
- Keep the project lightweight and maintainable

### Project Structure
```
cmd/main.go        # API entry point (main package)
auth/              # Authentication logic
db/                # Database access layer
handlers/          # HTTP handlers
middleware/        # HTTP middleware
web/              # HTML templates
schema.sql         # PostgreSQL schema
openapi.yaml       # OpenAPI contract (source of truth)
```

### Package Organization
- Keep packages small and focused (single word names)
- Use lowercase package names
- Each package should have a single responsibility
- Export types and functions that form the package's public API

### Imports
- Group imports into three blocks: stdlib, project packages, third-party
- Use `goimports` or `gofmt` to format automatically
- Avoid unused imports
- Keep import lists short and relevant

### Formatting
- Use `gofmt` for all code
- Maximum line length: 100 characters (soft limit)
- Use tabs for indentation
- Follow Effective Go guidelines

### Types and Interfaces
- Define types with clear purposes
- Use `er` suffix for interfaces (`Hasher`, `Repository`)
- Prefer concrete types when abstraction isn't needed
- Use structs to group related data

### Naming Conventions
- **Packages:** lowercase, single word
- **Interfaces:** `er` suffix (`Hasher`, `Repository`, `Validator`)
- **Constants:** `UPPER_SNAKE_CASE`
- **Functions:** `verbNoun` or `nounVerb`
- **Variables:** `camelCase`
- **Private members:** `camelCase` (unexported)
- **Public members:** `PascalCase` (exported)
- **Abbreviations:** remain uppercase (`HTTP`, `API`, `URL`, `JSON`, `SQL`)

### Function Signatures
- `context.Context` is always the first parameter
- Return errors as the last return value
- Use pointer receivers for methods that modify the struct
- Use value receivers for methods that don't modify the struct

### Error Handling
- Never ignore errors
- Wrap errors with context using `fmt.Errorf("operation: %w", err)`
- Define custom domain errors using `errors.New` or `fmt.Errorf`
- Use `errors.Is()` and `errors.As()` for error checking
- Log errors at appropriate levels (warn, error, fatal)
- Return meaningful error messages to clients

### HTTP Handlers
- Handler signature: `func(w http.ResponseWriter, r *http.Request) error`
- Always handle HTTP methods explicitly
- Use middleware for cross-cutting concerns (logging, auth, recovery)
- Parse JSON using `json.Decoder` instead of `json.Unmarshal`
- Always set `Content-Type` header: `w.Header().Set("Content-Type", "application/json")`
- Use correct HTTP status codes: 200, 201, 400, 401, 404, 500

### Database Access
- Always use parameterized queries (prevent SQL injection)
- Use prepared statements
- Handle `sql.ErrNoRows` explicitly (it's not an error, it's a missing result)
- Use transactions correctly: `Begin`, `defer Rollback`, `Commit` on success
- Apply context timeouts to all database operations
- Index frequently queried columns (see schema.sql)
- Never build SQL with string concatenation

### Security
- Use bcrypt for password hashing (cost factor 10)
- Hash all tokens before storing in database
- Validate and sanitize all inputs
- Enforce HTTPS in production
- Set appropriate security headers
- Never log or expose secrets/credentials

### OpenAPI Contract
- Contract-first development: `openapi.yaml` is the source of truth
- Keep implementation and documentation in sync
- Define all endpoints, schemas, and errors in the contract
- Use correct HTTP methods and status codes

### Testing
- Write table-driven tests for multiple test cases
- Cover both success and failure scenarios
- Use `httptest.ResponseRecorder` for HTTP handler tests
- Use `sqlmock` for database layer tests
- Tests must be fast and deterministic (no randomness, sleep, or external deps)
- Name tests descriptively: `TestFunctionName_SpecificScenario`

### HTML Templates
- Use `html/template` package
- Define templates in `web/` directory
- Use the base template pattern (see web/base.html)
- Follow the existing template structure and style
- Keep templates simple and readable

### Context Usage
- Always pass `context.Context` as the first parameter
- Set appropriate timeouts on contexts
- Never store contexts in structs
- Use context values sparingly (prefer explicit parameters)
- Check context cancellation in long-running operations

### Logging
- Use the standard library `log` package
- Log at appropriate levels (Debug, Info, Warn, Error)
- Include relevant context in log messages
- Don't log sensitive data (passwords, tokens, PII)
- Use structured logging when possible

## Development Workflow

1. Define endpoints and schemas in `openapi.yaml`
2. Implement database schema in `schema.sql`
3. Build Go API to match the OpenAPI contract
4. Write tests before implementation (TDD preferred)
5. Run tests: `go test ./...`
6. Run linter: `golangci-lint run`
7. Format code: `go fmt ./...`
8. Implement UI templates if needed

## Important Notes

- This is a framework-free project using only Go standard library
- Minimal dependencies = better security and maintainability
- Keep changes focused and atomic
- Always ensure tests pass before committing
- Review and update `openapi.yaml` when API changes are made
