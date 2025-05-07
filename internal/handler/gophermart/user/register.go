package user

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/riouske/gophermart/internal/model"
	"github.com/riouske/gophermart/internal/repository"
	"github.com/riouske/gophermart/internal/service"
)

type RegisterHandler struct {
	authService *service.AuthService
}

func NewRegisterHandler(authService *service.AuthService) *RegisterHandler {
	return &RegisterHandler{
		authService: authService,
	}
}

func (h *RegisterHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	var credentials model.UserCredentials
	if err := json.NewDecoder(r.Body).Decode(&credentials); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if credentials.Login == "" || credentials.Password == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	_, token, err := h.authService.Register(&credentials)
	if err != nil {
		if errors.Is(err, repository.ErrUserExists) {
			w.WriteHeader(http.StatusConflict)
			return
		}
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "auth_token",
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
	})
	w.Header().Set("Authorization", "Bearer "+token)
	w.WriteHeader(http.StatusOK)
}
