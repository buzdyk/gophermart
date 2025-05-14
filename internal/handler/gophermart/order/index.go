package order

import (
	"encoding/json"
	"net/http"

	"github.com/riouske/gophermart/internal/handler/middleware"
	"github.com/riouske/gophermart/internal/model"
	"github.com/riouske/gophermart/internal/repository"
)

// OrderResponse is a data transfer object for order list response
type OrderResponse struct {
	Number     string          `json:"number"`
	Status     model.OrderStatus `json:"status"`
	Accrual    *float64        `json:"accrual,omitempty"`
	UploadedAt string          `json:"uploaded_at"`
}

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

	// Convert to response format
	var responseOrders []OrderResponse
	for _, order := range orders {
		orderResp := OrderResponse{
			Number:     order.Number,
			Status:     order.Status,
			UploadedAt: order.UploadedAt.Format("2006-01-02T15:04:05-07:00"), // RFC3339 format
		}
		
		// Only include accrual if it's not zero
		if order.Accrual > 0 {
			accrual := order.Accrual
			orderResp.Accrual = &accrual
		}
		
		responseOrders = append(responseOrders, orderResp)
	}

	// Serialize orders to JSON
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(responseOrders); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}