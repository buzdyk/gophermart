package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/riouske/gophermart/internal/handler/middleware"
	"github.com/riouske/gophermart/internal/model"
	"github.com/riouske/gophermart/internal/tests"
)

func TestAuthMiddleware(t *testing.T) {
	// Setup
	authService, _, db := tests.SetupAuthService(t)
	defer db.Close()

	// Create a test user and generate a token
	user := &model.User{
		Login: "authtest",
	}
	_, token, err := authService.Register(&model.UserCredentials{
		Login:    user.Login,
		Password: "password123",
	})
	if err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}

	// Create a simple handler that will be wrapped by the auth middleware
	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID, ok := middleware.GetUserID(r.Context())
		if !ok {
			t.Error("User ID not found in context")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		if userID <= 0 {
			t.Errorf("Invalid user ID: %d", userID)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
	})

	// Create the auth middleware
	authMiddleware := middleware.Auth(authService)
	wrappedHandler := authMiddleware(testHandler)

	t.Run("ValidToken", func(t *testing.T) {
		req, err := http.NewRequest(http.MethodGet, "/protected", nil)
		if err != nil {
			t.Fatal(err)
		}
		req.Header.Set("Authorization", "Bearer "+token)

		rr := httptest.NewRecorder()
		wrappedHandler.ServeHTTP(rr, req)

		if status := rr.Code; status != http.StatusOK {
			t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
		}
	})

	t.Run("NoAuthHeader", func(t *testing.T) {
		req, err := http.NewRequest(http.MethodGet, "/protected", nil)
		if err != nil {
			t.Fatal(err)
		}

		rr := httptest.NewRecorder()
		wrappedHandler.ServeHTTP(rr, req)

		if status := rr.Code; status != http.StatusUnauthorized {
			t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusUnauthorized)
		}
	})

	t.Run("InvalidAuthFormat", func(t *testing.T) {
		// Missing "Bearer" prefix
		req, err := http.NewRequest(http.MethodGet, "/protected", nil)
		if err != nil {
			t.Fatal(err)
		}
		req.Header.Set("Authorization", token)

		rr := httptest.NewRecorder()
		wrappedHandler.ServeHTTP(rr, req)

		if status := rr.Code; status != http.StatusUnauthorized {
			t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusUnauthorized)
		}

		// Wrong format
		req, err = http.NewRequest(http.MethodGet, "/protected", nil)
		if err != nil {
			t.Fatal(err)
		}
		req.Header.Set("Authorization", "Basic "+token)

		rr = httptest.NewRecorder()
		wrappedHandler.ServeHTTP(rr, req)

		if status := rr.Code; status != http.StatusUnauthorized {
			t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusUnauthorized)
		}
	})

	t.Run("InvalidToken", func(t *testing.T) {
		req, err := http.NewRequest(http.MethodGet, "/protected", nil)
		if err != nil {
			t.Fatal(err)
		}
		req.Header.Set("Authorization", "Bearer invalid-token")

		rr := httptest.NewRecorder()
		wrappedHandler.ServeHTTP(rr, req)

		if status := rr.Code; status != http.StatusUnauthorized {
			t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusUnauthorized)
		}
	})

	t.Run("GetUserIDFromContext", func(t *testing.T) {
		// This test verifies the GetUserID helper function
		// Create a mock request with context
		req, err := http.NewRequest(http.MethodGet, "/", nil)
		if err != nil {
			t.Fatal(err)
		}
		
		// Create a mock context with a user ID
		ctx := middleware.WithUserID(req.Context(), 123)
		
		userID, ok := middleware.GetUserID(ctx)
		if !ok {
			t.Error("GetUserID returned not ok for valid context")
		}
		if userID != 123 {
			t.Errorf("GetUserID returned wrong ID: got %v want %v", userID, 123)
		}
		
		// Test with invalid context
		_, ok = middleware.GetUserID(req.Context())
		if ok {
			t.Error("GetUserID returned ok for invalid context")
		}
	})
}