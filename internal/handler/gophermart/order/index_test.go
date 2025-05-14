package order

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/riouske/gophermart/internal/handler/middleware"
	"github.com/riouske/gophermart/internal/model"
	"github.com/riouske/gophermart/internal/repository"
)

// MockOrderRepository for testing
type MockOrderListRepository struct {
	orders []*model.Order
}

func (m *MockOrderListRepository) Create(order *model.Order) error {
	return nil
}

func (m *MockOrderListRepository) GetByID(id int64) (*model.Order, error) {
	return nil, nil
}

func (m *MockOrderListRepository) GetByNumber(number string) (*model.Order, error) {
	return nil, nil
}

func (m *MockOrderListRepository) GetByUserID(userID int64) ([]*model.Order, error) {
	var userOrders []*model.Order
	for _, order := range m.orders {
		if order.UserID == userID {
			userOrders = append(userOrders, order)
		}
	}
	return userOrders, nil
}

func (m *MockOrderListRepository) UpdateStatus(id int64, status model.OrderStatus) error {
	return nil
}

func (m *MockOrderListRepository) UpdateAccrual(id int64, accrual float64, status model.OrderStatus) error {
	return nil
}

func TestIndexHandler_ServeHTTP(t *testing.T) {
	// Define test cases
	tests := []struct {
		name           string
		userID         int64
		orders         []*model.Order
		expectedStatus int
		checkResponse  bool
	}{
		{
			name:   "Success with orders",
			userID: 1,
			orders: []*model.Order{
				{
					ID:         1,
					UserID:     1,
					Number:     "9278923470",
					Status:     model.OrderStatusProcessed,
					Accrual:    500,
					UploadedAt: time.Now().Add(-time.Hour),
				},
				{
					ID:         2,
					UserID:     1,
					Number:     "12345678903",
					Status:     model.OrderStatusProcessing,
					Accrual:    0,
					UploadedAt: time.Now().Add(-2 * time.Hour),
				},
				{
					ID:         3,
					UserID:     1,
					Number:     "346436439",
					Status:     model.OrderStatusInvalid,
					Accrual:    0,
					UploadedAt: time.Now().Add(-24 * time.Hour),
				},
			},
			expectedStatus: http.StatusOK,
			checkResponse:  true,
		},
		{
			name:           "No orders",
			userID:         1,
			orders:         []*model.Order{},
			expectedStatus: http.StatusNoContent,
			checkResponse:  false,
		},
		{
			name:   "Orders for different user",
			userID: 1,
			orders: []*model.Order{
				{
					ID:     4,
					UserID: 2,
					Number: "123456789",
					Status: model.OrderStatusNew,
				},
			},
			expectedStatus: http.StatusNoContent,
			checkResponse:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock repository with test data
			mockImpl := &MockOrderListRepository{
				orders: tt.orders,
			}
			mockRepo := &repository.OrderRepository{
				Impl: mockImpl,
			}

			// Create the handler with the mock repository
			handler := NewIndexHandler(mockRepo)

			// Create a request
			req := httptest.NewRequest(http.MethodGet, "/api/user/orders", nil)
			req = req.WithContext(middleware.WithUserID(context.Background(), tt.userID))

			// Create a response recorder
			rr := httptest.NewRecorder()

			// Call the handler
			handler.ServeHTTP(rr, req)

			// Check the status code
			if status := rr.Code; status != tt.expectedStatus {
				t.Errorf("handler returned wrong status code: got %v want %v", status, tt.expectedStatus)
			}

			// If we expect a response, check its format
			if tt.checkResponse {
				var response []OrderResponse
				if err := json.NewDecoder(rr.Body).Decode(&response); err != nil {
					t.Errorf("Failed to decode response: %v", err)
				}

				// Check that we got the expected number of orders
				expectedUserOrders := 0
				for _, order := range tt.orders {
					if order.UserID == tt.userID {
						expectedUserOrders++
					}
				}
				if len(response) != expectedUserOrders {
					t.Errorf("Expected %d orders, got %d", expectedUserOrders, len(response))
				}

				// Check that orders are sorted by uploaded_at desc
				if len(response) > 1 {
					for i := 0; i < len(response)-1; i++ {
						t1, err1 := time.Parse("2006-01-02T15:04:05-07:00", response[i].UploadedAt)
						t2, err2 := time.Parse("2006-01-02T15:04:05-07:00", response[i+1].UploadedAt)
						if err1 != nil || err2 != nil {
							t.Errorf("Failed to parse dates in response")
						}
						if t1.Before(t2) {
							t.Errorf("Orders not sorted by uploaded_at desc")
						}
					}
				}

				// Check that accrual is only present when it should be
				for _, resp := range response {
					var found bool
					var expectedAccrual *float64

					for _, order := range tt.orders {
						if order.Number == resp.Number {
							found = true
							if order.Accrual > 0 {
								a := order.Accrual
								expectedAccrual = &a
							} else {
								expectedAccrual = nil
							}
							break
						}
					}

					if !found {
						t.Errorf("Response contains unexpected order: %s", resp.Number)
					}

					if (expectedAccrual == nil && resp.Accrual != nil) ||
						(expectedAccrual != nil && resp.Accrual == nil) {
						t.Errorf("Accrual mismatch for order %s: expected %v, got %v", resp.Number, expectedAccrual, resp.Accrual)
					}

					if expectedAccrual != nil && resp.Accrual != nil && *expectedAccrual != *resp.Accrual {
						t.Errorf("Accrual value mismatch for order %s: expected %v, got %v", resp.Number, *expectedAccrual, *resp.Accrual)
					}
				}
			}
		})
	}
}