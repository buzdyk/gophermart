package user_test

import (
	"net/http"
	"testing"

	"github.com/riouske/gophermart/internal/handler/gophermart/user"
	"github.com/riouske/gophermart/internal/model"
	"github.com/riouske/gophermart/internal/tests"
)

func TestRegisterHandler(t *testing.T) {
	// Setup
	authService, _, db := tests.SetupAuthService(t)
	defer db.Close()

	handler := user.NewRegisterHandler(authService)

	t.Run("RegisterValidUser", func(t *testing.T) {
		credentials := model.UserCredentials{
			Login:    "testuser",
			Password: "password123",
		}

		rr := tests.MakeRequest(t, http.MethodPost, "/api/user/register", credentials, handler)

		// Check response
		if status := rr.Code; status != http.StatusOK {
			t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusOK)
		}

		// Check if token is returned
		token := tests.ExtractAuthToken(rr)
		if token == "" {
			t.Error("expected auth token in response, got none")
		}

		// Check if cookie is set
		cookies := rr.Result().Cookies()
		var authCookie *http.Cookie
		for _, cookie := range cookies {
			if cookie.Name == "auth_token" {
				authCookie = cookie
				break
			}
		}
		if authCookie == nil {
			t.Error("auth_token cookie not set")
		} else if authCookie.Value == "" {
			t.Error("auth_token cookie value is empty")
		}
	})

	t.Run("RegisterDuplicateUser", func(t *testing.T) {
		credentials := model.UserCredentials{
			Login:    "duplicate",
			Password: "password123",
		}

		// Register user first time
		rr := tests.MakeRequest(t, http.MethodPost, "/api/user/register", credentials, handler)
		if status := rr.Code; status != http.StatusOK {
			t.Errorf("first registration: handler returned wrong status code: got %v want %v", status, http.StatusOK)
		}

		// Try to register same user again
		rr = tests.MakeRequest(t, http.MethodPost, "/api/user/register", credentials, handler)
		if status := rr.Code; status != http.StatusConflict {
			t.Errorf("duplicate registration: handler returned wrong status code: got %v want %v", status, http.StatusConflict)
		}
	})

	t.Run("RegisterWithEmptyCredentials", func(t *testing.T) {
		// Empty login
		credentials := model.UserCredentials{
			Login:    "",
			Password: "password123",
		}
		rr := tests.MakeRequest(t, http.MethodPost, "/api/user/register", credentials, handler)
		if status := rr.Code; status != http.StatusBadRequest {
			t.Errorf("empty login: handler returned wrong status code: got %v want %v", status, http.StatusBadRequest)
		}

		// Empty password
		credentials = model.UserCredentials{
			Login:    "testuser2",
			Password: "",
		}
		rr = tests.MakeRequest(t, http.MethodPost, "/api/user/register", credentials, handler)
		if status := rr.Code; status != http.StatusBadRequest {
			t.Errorf("empty password: handler returned wrong status code: got %v want %v", status, http.StatusBadRequest)
		}
	})

	t.Run("RegisterWithInvalidMethod", func(t *testing.T) {
		rr := tests.MakeRequest(t, http.MethodGet, "/api/user/register", nil, handler)
		if status := rr.Code; status != http.StatusMethodNotAllowed {
			t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusMethodNotAllowed)
		}
	})

	t.Run("RegisterWithInvalidJSON", func(t *testing.T) {
		invalidJSON := `{"login": "testuser", "password": "password123"`
		req, err := http.NewRequest(http.MethodPost, "/api/user/register", tests.NewStringReader(invalidJSON))
		if err != nil {
			t.Fatal(err)
		}
		req.Header.Set("Content-Type", "application/json")

		rr := tests.NewRecorder()
		handler.ServeHTTP(rr, req)

		if status := rr.Code; status != http.StatusBadRequest {
			t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusBadRequest)
		}
	})
}