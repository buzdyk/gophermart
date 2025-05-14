package order

import (
	"errors"
	"io"
	"net/http"
	"strings"

	"github.com/riouske/gophermart/internal/handler/middleware"
	"github.com/riouske/gophermart/internal/model"
	"github.com/riouske/gophermart/internal/repository"
	"github.com/riouske/gophermart/internal/util"
)

type CreateHandler struct {
	orderRepo *repository.OrderRepository
}

func NewCreateHandler(orderRepo *repository.OrderRepository) *CreateHandler {
	return &CreateHandler{
		orderRepo: orderRepo,
	}
}

func (h *CreateHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	// Get user ID from context
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	// Read the order number from the request body
	body, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	orderNumber := strings.TrimSpace(string(body))
	if orderNumber == "" {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Validate the order number using Luhn algorithm
	if !util.ValidateLuhn(orderNumber) {
		w.WriteHeader(http.StatusUnprocessableEntity)
		return
	}

	// Create a new order
	order := &model.Order{
		UserID: userID,
		Number: orderNumber,
		Status: model.OrderStatusNew,
	}

	// Attempt to save the order
	err = h.orderRepo.Create(order)
	if err != nil {
		switch {
		case errors.Is(err, repository.ErrOrderExists):
			// Order already exists for this user
			w.WriteHeader(http.StatusOK)
		case errors.Is(err, repository.ErrOrderExistsForUser):
			// Order already exists for another user
			w.WriteHeader(http.StatusConflict)
		default:
			// Internal server error
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}

	// Order successfully added
	w.WriteHeader(http.StatusAccepted)
}