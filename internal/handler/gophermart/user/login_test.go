package user_test

import (
	"net/http"
	"testing"

	"github.com/riouske/gophermart/internal/handler/gophermart/user"
	"github.com/riouske/gophermart/internal/model"
	"github.com/riouske/gophermart/internal/tests"
)

func TestLoginHandler(t *testing.T) {
	// Setup
	authService, _, db := tests.SetupAuthService(t)
	defer db.Close()

	registerHandler := user.NewRegisterHandler(authService)
	loginHandler := user.NewLoginHandler(authService)

	// Pre-register a user for testing login
	testUser := model.UserCredentials{
		Login:    "logintest",
		Password: "password123",
	}
	tests.MakeRequest(t, http.MethodPost, "/api/user/register", testUser, registerHandler)

	t.Run("LoginValidUser", func(t *testing.T) {
		credentials := model.UserCredentials{
			Login:    "logintest",
			Password: "password123",
		}

		rr := tests.MakeRequest(t, http.MethodPost, "/api/user/login", credentials, loginHandler)

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

	t.Run("LoginNonExistentUser", func(t *testing.T) {
		credentials := model.UserCredentials{
			Login:    "nonexistent",
			Password: "password123",
		}

		rr := tests.MakeRequest(t, http.MethodPost, "/api/user/login", credentials, loginHandler)
		if status := rr.Code; status != http.StatusUnauthorized {
			t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusUnauthorized)
		}
	})

	t.Run("LoginWithWrongPassword", func(t *testing.T) {
		credentials := model.UserCredentials{
			Login:    "logintest",
			Password: "wrongpassword",
		}

		rr := tests.MakeRequest(t, http.MethodPost, "/api/user/login", credentials, loginHandler)
		if status := rr.Code; status != http.StatusUnauthorized {
			t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusUnauthorized)
		}
	})

	t.Run("LoginWithEmptyCredentials", func(t *testing.T) {
		// Empty login
		credentials := model.UserCredentials{
			Login:    "",
			Password: "password123",
		}
		rr := tests.MakeRequest(t, http.MethodPost, "/api/user/login", credentials, loginHandler)
		if status := rr.Code; status != http.StatusBadRequest {
			t.Errorf("empty login: handler returned wrong status code: got %v want %v", status, http.StatusBadRequest)
		}

		// Empty password
		credentials = model.UserCredentials{
			Login:    "logintest",
			Password: "",
		}
		rr = tests.MakeRequest(t, http.MethodPost, "/api/user/login", credentials, loginHandler)
		if status := rr.Code; status != http.StatusBadRequest {
			t.Errorf("empty password: handler returned wrong status code: got %v want %v", status, http.StatusBadRequest)
		}
	})

	t.Run("LoginWithInvalidMethod", func(t *testing.T) {
		rr := tests.MakeRequest(t, http.MethodGet, "/api/user/login", nil, loginHandler)
		if status := rr.Code; status != http.StatusMethodNotAllowed {
			t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusMethodNotAllowed)
		}
	})

	t.Run("LoginWithInvalidJSON", func(t *testing.T) {
		invalidJSON := `{"login": "testuser", "password": "password123"`
		req, err := http.NewRequest(http.MethodPost, "/api/user/login", tests.NewStringReader(invalidJSON))
		if err != nil {
			t.Fatal(err)
		}
		req.Header.Set("Content-Type", "application/json")

		rr := tests.NewRecorder()
		loginHandler.ServeHTTP(rr, req)

		if status := rr.Code; status != http.StatusBadRequest {
			t.Errorf("handler returned wrong status code: got %v want %v", status, http.StatusBadRequest)
		}
	})
}