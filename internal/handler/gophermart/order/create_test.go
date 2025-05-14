package order

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/riouske/gophermart/internal/handler/middleware"
	"github.com/riouske/gophermart/internal/model"
	"github.com/riouske/gophermart/internal/repository"
)

// MockOrderRepository is a test mock for OrderRepository
type MockOrderRepository struct {
	orders map[string]*model.Order
}

func NewMockOrderRepository() *repository.OrderRepository {
	mock := &MockOrderRepository{
		orders: make(map[string]*model.Order),
	}
	// Type assertion to ensure mock implements the interface
	var _ repository.OrderRepositoryInterface = mock
	return &repository.OrderRepository{
		Impl: mock,
	}
}

func (m *MockOrderRepository) Create(order *model.Order) error {
	if existing, ok := m.orders[order.Number]; ok {
		if existing.UserID == order.UserID {
			return repository.ErrOrderExists
		}
		return repository.ErrOrderExistsForUser
	}

	m.orders[order.Number] = order
	return nil
}

func (m *MockOrderRepository) GetByUserID(userID int64) ([]*model.Order, error) {
	var result []*model.Order
	for _, order := range m.orders {
		if order.UserID == userID {
			result = append(result, order)
		}
	}
	return result, nil
}

func (m *MockOrderRepository) GetByNumber(number string) (*model.Order, error) {
	if order, ok := m.orders[number]; ok {
		return order, nil
	}
	return nil, repository.ErrOrderNotFound
}

func (m *MockOrderRepository) GetByID(id int64) (*model.Order, error) {
	for _, order := range m.orders {
		if order.ID == id {
			return order, nil
		}
	}
	return nil, repository.ErrOrderNotFound
}

func (m *MockOrderRepository) UpdateStatus(id int64, status model.OrderStatus) error {
	return nil
}

func (m *MockOrderRepository) UpdateAccrual(id int64, accrual float64, status model.OrderStatus) error {
	return nil
}

func TestCreateHandler_ServeHTTP(t *testing.T) {
	tests := []struct {
		name           string
		orderNumber    string
		existingOrders map[string]*model.Order
		userID         int64
		wantStatus     int
	}{
		{
			name:        "Valid order number, success",
			orderNumber: "79927398713", // Valid Luhn
			userID:      1,
			wantStatus:  http.StatusAccepted,
		},
		{
			name:        "Invalid order number",
			orderNumber: "123456", // Invalid Luhn
			userID:      1,
			wantStatus:  http.StatusUnprocessableEntity,
		},
		{
			name:        "Order already exists for this user",
			orderNumber: "79927398713", // Valid Luhn
			existingOrders: map[string]*model.Order{
				"79927398713": {
					UserID: 1,
					Number: "79927398713",
				},
			},
			userID:     1,
			wantStatus: http.StatusOK,
		},
		{
			name:        "Order exists for another user",
			orderNumber: "79927398713", // Valid Luhn
			existingOrders: map[string]*model.Order{
				"79927398713": {
					UserID: 2,
					Number: "79927398713",
				},
			},
			userID:     1,
			wantStatus: http.StatusConflict,
		},
		{
			name:        "Empty order number",
			orderNumber: "",
			userID:      1,
			wantStatus:  http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a mock repo
			mockImpl := &MockOrderRepository{
				orders: make(map[string]*model.Order),
			}
			if tt.existingOrders != nil {
				mockImpl.orders = tt.existingOrders
			}
			
			mockRepo := &repository.OrderRepository{
				Impl: mockImpl,
			}

			handler := NewCreateHandler(mockRepo)

			// Create a request
			req := httptest.NewRequest(http.MethodPost, "/api/user/orders", strings.NewReader(tt.orderNumber))
			req = req.WithContext(middleware.WithUserID(context.Background(), tt.userID))

			// Create a response recorder
			rr := httptest.NewRecorder()

			// Serve the request
			handler.ServeHTTP(rr, req)

			// Check the status code
			if status := rr.Code; status != tt.wantStatus {
				t.Errorf("handler returned wrong status code: got %v want %v", status, tt.wantStatus)
			}
		})
	}
}