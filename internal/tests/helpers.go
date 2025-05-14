package tests

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/riouske/gophermart/internal/config"
	"github.com/riouske/gophermart/internal/db"
	"github.com/riouske/gophermart/internal/repository"
	"github.com/riouske/gophermart/internal/service"
)

// TestDB creates a database connection for testing
func TestDB(t *testing.T) *sql.DB {
	t.Helper()
	// Use test database configuration
	dsn := "postgres://postgres:secret@localhost:5432/gophermart_test?sslmode=disable"
	
	// Skip database tests if test database is not available
	database, err := db.NewDB(dsn)
	if err != nil {
		t.Skipf("Skipping test that requires database: %v", err)
		return nil
	}
	return database
}

// CleanupDB truncates all test tables
func CleanupDB(t *testing.T, db *sql.DB) {
	t.Helper()
	_, err := db.Exec("TRUNCATE users CASCADE")
	if err != nil {
		t.Fatalf("Failed to clean up database: %v", err)
	}
}

// TestConfig returns a test configuration
func TestConfig() *config.Config {
	return &config.Config{
		DatabaseURI:       "postgres://postgres:secret@localhost:5432/gophermart_test?sslmode=disable",
		ServerAddress:     ":9091",
		AccrualSystemAddr: "http://localhost:8080",
		JWTSecretKey:      "test-secret-key",
	}
}

// SetupUserRepo sets up a clean user repository for testing
func SetupUserRepo(t *testing.T) (*repository.UserRepository, *sql.DB) {
	t.Helper()
	db := TestDB(t)
	CleanupDB(t, db)
	return repository.NewUserRepository(db), db
}

// SetupAuthService sets up the auth service for testing
func SetupAuthService(t *testing.T) (*service.AuthService, *repository.UserRepository, *sql.DB) {
	t.Helper()
	userRepo, db := SetupUserRepo(t)
	cfg := TestConfig()
	authService := service.NewAuthService(userRepo, cfg.JWTSecretKey)
	return authService, userRepo, db
}

// MakeRequest is a helper to make HTTP requests in tests
func MakeRequest(t *testing.T, method, url string, body interface{}, handler http.Handler) *httptest.ResponseRecorder {
	t.Helper()
	var reqBody io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			t.Fatalf("Failed to marshal request body: %v", err)
		}
		reqBody = bytes.NewBuffer(jsonBody)
	}

	req, err := http.NewRequest(method, url, reqBody)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	return rr
}

// ParseResponseBody parses the response body into the given interface
func ParseResponseBody(t *testing.T, rr *httptest.ResponseRecorder, v interface{}) {
	t.Helper()
	err := json.Unmarshal(rr.Body.Bytes(), v)
	if err != nil {
		t.Fatalf("Failed to parse response body: %v", err)
	}
}

// ExtractAuthToken extracts the auth token from the response headers
func ExtractAuthToken(rr *httptest.ResponseRecorder) string {
	authHeader := rr.Header().Get("Authorization")
	if authHeader == "" {
		return ""
	}
	// Format is "Bearer <token>"
	return authHeader[7:]
}

// NewStringReader creates an io.Reader from a string
func NewStringReader(s string) io.Reader {
	return strings.NewReader(s)
}

// NewRecorder creates a new ResponseRecorder
func NewRecorder() *httptest.ResponseRecorder {
	return httptest.NewRecorder()
}