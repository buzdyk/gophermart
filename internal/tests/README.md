# Testing Guidelines for GopherMart

This document outlines how to run tests for the GopherMart application, specifically for the authentication endpoints.

## Prerequisites

- PostgreSQL running locally or in Docker
- Test database created: `gophermart_test`

## Test Database Setup

Before running tests, ensure you have a test database created:

```bash
createdb gophermart_test
```

Or with Docker:

```bash
docker exec -it [postgres-container-name] createdb -U postgres gophermart_test
```

Then run the migrations:

```bash
go run migrations/migrate.go -database "postgres://postgres:secret@localhost:5432/gophermart_test?sslmode=disable" -path ./migrations up
```

## Running the Tests

Run all tests with:

```bash
go test ./...
```

Run specific test packages:

```bash
go test ./internal/handler/gophermart/user/...
go test ./internal/handler/middleware/...
```

Run individual test files:

```bash
go test ./internal/handler/gophermart/user/register_test.go
go test ./internal/handler/gophermart/user/login_test.go
go test ./internal/handler/middleware/auth_test.go
```

## Test Structure

The tests are structured as follows:

1. **Test Helpers** (`internal/tests/helpers.go`):
   - Common utilities for setting up test database connections
   - Helpers for making HTTP requests and validating responses
   - Utilities for creating test fixtures

2. **Auth Endpoint Tests**:
   - `register_test.go`: Tests for user registration endpoint
   - `login_test.go`: Tests for user login endpoint
   - `auth_test.go`: Tests for the authentication middleware

## Test Coverage

The current tests cover:

- User registration (success and various failure cases)
- User login (success and various failure cases)
- Authentication middleware (token validation, extraction, context management)

## Adding New Tests

When adding new tests, follow these guidelines:

1. Use the helper functions in `internal/tests/helpers.go`
2. Clean up the database before each test
3. Use subtests for different test cases
4. Test both success and failure cases
5. Verify all expected HTTP status codes and responses

## Mocking Dependencies

The tests currently use a real PostgreSQL database. If you want to add mock implementations for faster tests:

1. Create interfaces for repositories and services
2. Implement mock versions for testing
3. Update the test helpers to support both real and mock implementations