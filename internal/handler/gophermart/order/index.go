package order

import (
	"encoding/json"
	"net/http"

	"github.com/riouske/gophermart/internal/handler/middleware"
	"github.com/riouske/gophermart/internal/repository"
)

type IndexHandler struct {
	orderRepo *repository.OrderRepository
}

func NewIndexHandler(orderRepo *repository.OrderRepository) *IndexHandler {
	return &IndexHandler{
		orderRepo: orderRepo,
	}
}

func (h *IndexHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	// Get user ID from context
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	// Get all orders for the user
	orders, err := h.orderRepo.GetByUserID(userID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Return 204 if no orders found
	if len(orders) == 0 {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	// Serialize orders to JSON
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(orders); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}