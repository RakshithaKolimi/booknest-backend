package controller

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"booknest/internal/domain"
)

type mockOrderServiceController struct {
	checkoutFunc               func(ctx context.Context, userID uuid.UUID, input domain.CheckoutInput) (domain.OrderView, error)
	confirmPaymentFunc         func(ctx context.Context, userID uuid.UUID, input domain.PaymentConfirmInput) (domain.OrderView, error)
	cancelOrderFunc            func(ctx context.Context, userID uuid.UUID, input domain.OrderCancelInput) (domain.OrderView, error)
	adminUpdateOrderStatusFunc func(ctx context.Context, input domain.AdminOrderStatusUpdateInput) (domain.OrderView, error)
	listUserOrdersFunc         func(ctx context.Context, userID uuid.UUID, limit, offset int) ([]domain.OrderView, error)
	listAllOrdersFunc          func(ctx context.Context, limit, offset int) ([]domain.OrderView, error)
}

func (m *mockOrderServiceController) Checkout(ctx context.Context, userID uuid.UUID, input domain.CheckoutInput) (domain.OrderView, error) {
	if m.checkoutFunc != nil {
		return m.checkoutFunc(ctx, userID, input)
	}
	return domain.OrderView{}, errors.New("not implemented")
}
func (m *mockOrderServiceController) ConfirmPayment(ctx context.Context, userID uuid.UUID, input domain.PaymentConfirmInput) (domain.OrderView, error) {
	if m.confirmPaymentFunc != nil {
		return m.confirmPaymentFunc(ctx, userID, input)
	}
	return domain.OrderView{}, errors.New("not implemented")
}
func (m *mockOrderServiceController) CancelOrder(ctx context.Context, userID uuid.UUID, input domain.OrderCancelInput) (domain.OrderView, error) {
	if m.cancelOrderFunc != nil {
		return m.cancelOrderFunc(ctx, userID, input)
	}
	return domain.OrderView{}, errors.New("not implemented")
}
func (m *mockOrderServiceController) AdminUpdateOrderStatus(ctx context.Context, input domain.AdminOrderStatusUpdateInput) (domain.OrderView, error) {
	if m.adminUpdateOrderStatusFunc != nil {
		return m.adminUpdateOrderStatusFunc(ctx, input)
	}
	return domain.OrderView{}, errors.New("not implemented")
}
func (m *mockOrderServiceController) ListUserOrders(ctx context.Context, userID uuid.UUID, limit, offset int) ([]domain.OrderView, error) {
	if m.listUserOrdersFunc != nil {
		return m.listUserOrdersFunc(ctx, userID, limit, offset)
	}
	return []domain.OrderView{}, nil
}
func (m *mockOrderServiceController) ListAllOrders(ctx context.Context, limit, offset int) ([]domain.OrderView, error) {
	if m.listAllOrdersFunc != nil {
		return m.listAllOrdersFunc(ctx, limit, offset)
	}
	return []domain.OrderView{}, nil
}

func TestOrderControllerCheckoutConfirmAndCancel(t *testing.T) {
	gin.SetMode(gin.TestMode)
	userID := uuid.New()
	orderID := uuid.New()
	calledCheckout := false
	calledConfirm := false
	calledCancel := false

	svc := &mockOrderServiceController{
		checkoutFunc: func(ctx context.Context, gotUserID uuid.UUID, input domain.CheckoutInput) (domain.OrderView, error) {
			calledCheckout = true
			if gotUserID != userID || input.PaymentMethod != domain.PaymentCOD {
				t.Fatalf("unexpected checkout input")
			}
			return domain.OrderView{Order: domain.Order{ID: orderID}}, nil
		},
		confirmPaymentFunc: func(ctx context.Context, gotUserID uuid.UUID, input domain.PaymentConfirmInput) (domain.OrderView, error) {
			calledConfirm = true
			if gotUserID != userID || input.OrderID != orderID || !input.Success {
				t.Fatalf("unexpected confirm input")
			}
			return domain.OrderView{Order: domain.Order{ID: orderID}}, nil
		},
		cancelOrderFunc: func(ctx context.Context, gotUserID uuid.UUID, input domain.OrderCancelInput) (domain.OrderView, error) {
			calledCancel = true
			if gotUserID != userID || input.OrderID != orderID || input.CancellationReason != "Changed my mind" {
				t.Fatalf("unexpected cancel input")
			}
			return domain.OrderView{Order: domain.Order{ID: orderID}}, nil
		},
	}
	ctl := NewOrderController(svc).(*orderController)

	checkoutBody, _ := json.Marshal(domain.CheckoutInput{PaymentMethod: domain.PaymentCOD})
	cw := httptest.NewRecorder()
	cc, _ := gin.CreateTestContext(cw)
	cc.Set("user_id", userID.String())
	cc.Request = httptest.NewRequest(http.MethodPost, "/orders/checkout", bytes.NewBuffer(checkoutBody))
	cc.Request.Header.Set("Content-Type", "application/json")
	ctl.Checkout(cc)
	if cw.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", cw.Code)
	}

	confirmBody, _ := json.Marshal(domain.PaymentConfirmInput{OrderID: orderID, Success: true})
	rw := httptest.NewRecorder()
	rc, _ := gin.CreateTestContext(rw)
	rc.Set("user_id", userID.String())
	rc.Request = httptest.NewRequest(http.MethodPost, "/orders/confirm", bytes.NewBuffer(confirmBody))
	rc.Request.Header.Set("Content-Type", "application/json")
	ctl.ConfirmPayment(rc)
	if rw.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rw.Code)
	}

	cancelBody, _ := json.Marshal(domain.OrderCancelInput{OrderID: orderID, CancellationReason: "Changed my mind"})
	xw := httptest.NewRecorder()
	xc, _ := gin.CreateTestContext(xw)
	xc.Set("user_id", userID.String())
	xc.Request = httptest.NewRequest(http.MethodPost, "/orders/cancel", bytes.NewBuffer(cancelBody))
	xc.Request.Header.Set("Content-Type", "application/json")
	ctl.CancelOrder(xc)
	if xw.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", xw.Code)
	}

	if !calledCheckout || !calledConfirm || !calledCancel {
		t.Fatalf("expected checkout, confirm, and cancel handlers to call service")
	}
}

func TestOrderControllerListEndpoints(t *testing.T) {
	gin.SetMode(gin.TestMode)
	userID := uuid.New()
	svc := &mockOrderServiceController{
		listUserOrdersFunc: func(ctx context.Context, gotUserID uuid.UUID, limit, offset int) ([]domain.OrderView, error) {
			if gotUserID != userID || limit != 3 || offset != 1 {
				t.Fatalf("unexpected user list params")
			}
			return []domain.OrderView{}, nil
		},
		listAllOrdersFunc: func(ctx context.Context, limit, offset int) ([]domain.OrderView, error) {
			if limit != 4 || offset != 2 {
				t.Fatalf("unexpected admin list params")
			}
			return []domain.OrderView{}, nil
		},
	}
	ctl := NewOrderController(svc).(*orderController)

	uw := httptest.NewRecorder()
	uc, _ := gin.CreateTestContext(uw)
	uc.Set("user_id", userID.String())
	uc.Request = httptest.NewRequest(http.MethodGet, "/orders?limit=3&offset=1", nil)
	ctl.ListMyOrders(uc)
	if uw.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", uw.Code)
	}

	aw := httptest.NewRecorder()
	ac, _ := gin.CreateTestContext(aw)
	ac.Request = httptest.NewRequest(http.MethodGet, "/admin/orders?limit=4&offset=2", nil)
	ctl.ListAllOrders(ac)
	if aw.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", aw.Code)
	}
}

func TestOrderControllerAdminUpdateStatus(t *testing.T) {
	gin.SetMode(gin.TestMode)
	orderID := uuid.New()
	called := false
	svc := &mockOrderServiceController{
		adminUpdateOrderStatusFunc: func(ctx context.Context, input domain.AdminOrderStatusUpdateInput) (domain.OrderView, error) {
			called = true
			if input.OrderID != orderID || input.Status != domain.OrderCancelled || input.CancellationReason != "Out of stock" {
				t.Fatalf("unexpected admin status input")
			}
			return domain.OrderView{Order: domain.Order{ID: orderID}}, nil
		},
	}
	ctl := NewOrderController(svc).(*orderController)

	body, _ := json.Marshal(domain.AdminOrderStatusUpdateInput{
		OrderID:            orderID,
		Status:             domain.OrderCancelled,
		CancellationReason: "Out of stock",
	})
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPut, "/admin/orders/status", bytes.NewBuffer(body))
	c.Request.Header.Set("Content-Type", "application/json")

	ctl.AdminUpdateOrderStatus(c)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	if !called {
		t.Fatalf("expected admin update handler to call service")
	}
}

func TestOrderControllerAdminUpdatePaymentStatus(t *testing.T) {
	gin.SetMode(gin.TestMode)
	orderID := uuid.New()
	refunded := domain.PaymentRefunded
	called := false
	svc := &mockOrderServiceController{
		adminUpdateOrderStatusFunc: func(ctx context.Context, input domain.AdminOrderStatusUpdateInput) (domain.OrderView, error) {
			called = true
			if input.OrderID != orderID || input.PaymentStatus == nil || *input.PaymentStatus != domain.PaymentRefunded {
				t.Fatalf("unexpected admin payment input")
			}
			return domain.OrderView{Order: domain.Order{ID: orderID}}, nil
		},
	}
	ctl := NewOrderController(svc).(*orderController)

	body, _ := json.Marshal(domain.AdminOrderStatusUpdateInput{
		OrderID:       orderID,
		PaymentStatus: &refunded,
	})
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPut, "/admin/orders/status", bytes.NewBuffer(body))
	c.Request.Header.Set("Content-Type", "application/json")

	ctl.AdminUpdateOrderStatus(c)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	if !called {
		t.Fatalf("expected admin payment update handler to call service")
	}
}

func TestOrderControllerUnauthorized(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ctl := NewOrderController(&mockOrderServiceController{}).(*orderController)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/orders", nil)

	ctl.ListMyOrders(c)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", w.Code)
	}
}

func TestOrderControllerCancelOrderUnauthorized(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ctl := NewOrderController(&mockOrderServiceController{}).(*orderController)

	body, _ := json.Marshal(domain.OrderCancelInput{OrderID: uuid.New(), CancellationReason: "Changed my mind"})
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/orders/cancel", bytes.NewBuffer(body))
	c.Request.Header.Set("Content-Type", "application/json")

	ctl.CancelOrder(c)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", w.Code)
	}
}

func TestOrderControllerCheckoutAndConfirmErrorPaths(t *testing.T) {
	gin.SetMode(gin.TestMode)
	userID := uuid.New()
	orderID := uuid.New()
	ctl := NewOrderController(&mockOrderServiceController{
		checkoutFunc: func(ctx context.Context, userID uuid.UUID, input domain.CheckoutInput) (domain.OrderView, error) {
			return domain.OrderView{}, errors.New("checkout failed")
		},
		confirmPaymentFunc: func(ctx context.Context, userID uuid.UUID, input domain.PaymentConfirmInput) (domain.OrderView, error) {
			return domain.OrderView{}, errors.New("confirm failed")
		},
	}).(*orderController)

	tests := []struct {
		name   string
		method func(*gin.Context)
		setup  func(*gin.Context)
		body   string
		code   int
	}{
		{name: "checkout unauthorized", method: ctl.Checkout, body: `{"payment_method":"COD"}`, code: http.StatusUnauthorized},
		{name: "checkout bad json", method: ctl.Checkout, setup: func(c *gin.Context) { c.Set("user_id", userID.String()) }, body: `{`, code: http.StatusBadRequest},
		{name: "checkout service error", method: ctl.Checkout, setup: func(c *gin.Context) { c.Set("user_id", userID.String()) }, body: `{"payment_method":"COD"}`, code: http.StatusBadRequest},
		{name: "confirm unauthorized", method: ctl.ConfirmPayment, body: `{"order_id":"` + orderID.String() + `","success":true}`, code: http.StatusUnauthorized},
		{name: "confirm bad json", method: ctl.ConfirmPayment, setup: func(c *gin.Context) { c.Set("user_id", userID.String()) }, body: `{`, code: http.StatusBadRequest},
		{name: "confirm service error", method: ctl.ConfirmPayment, setup: func(c *gin.Context) { c.Set("user_id", userID.String()) }, body: `{"order_id":"` + orderID.String() + `","success":true}`, code: http.StatusBadRequest},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			if tc.setup != nil {
				tc.setup(c)
			}
			c.Request = httptest.NewRequest(http.MethodPost, "/test", bytes.NewBufferString(tc.body))
			c.Request.Header.Set("Content-Type", "application/json")
			tc.method(c)
			if w.Code != tc.code {
				t.Fatalf("expected %d, got %d", tc.code, w.Code)
			}
		})
	}
}

func TestOrderControllerCancelOrderBadRequest(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ctl := NewOrderController(&mockOrderServiceController{}).(*orderController)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set("user_id", uuid.New().String())
	c.Request = httptest.NewRequest(http.MethodPost, "/orders/cancel", bytes.NewBufferString("{"))
	c.Request.Header.Set("Content-Type", "application/json")

	ctl.CancelOrder(c)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestOrderControllerCancelOrderServiceError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := &mockOrderServiceController{
		cancelOrderFunc: func(ctx context.Context, userID uuid.UUID, input domain.OrderCancelInput) (domain.OrderView, error) {
			return domain.OrderView{}, errors.New("cannot cancel order")
		},
	}
	ctl := NewOrderController(svc).(*orderController)

	body, _ := json.Marshal(domain.OrderCancelInput{OrderID: uuid.New(), CancellationReason: "Changed my mind"})
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Set("user_id", uuid.New().String())
	c.Request = httptest.NewRequest(http.MethodPost, "/orders/cancel", bytes.NewBuffer(body))
	c.Request.Header.Set("Content-Type", "application/json")

	ctl.CancelOrder(c)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestOrderControllerAdminUpdateStatusBadRequest(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ctl := NewOrderController(&mockOrderServiceController{}).(*orderController)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPut, "/admin/orders/status", bytes.NewBufferString("{"))
	c.Request.Header.Set("Content-Type", "application/json")

	ctl.AdminUpdateOrderStatus(c)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestOrderControllerAdminUpdateStatusServiceError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := &mockOrderServiceController{
		adminUpdateOrderStatusFunc: func(ctx context.Context, input domain.AdminOrderStatusUpdateInput) (domain.OrderView, error) {
			return domain.OrderView{}, errors.New("invalid transition")
		},
	}
	ctl := NewOrderController(svc).(*orderController)

	body, _ := json.Marshal(domain.AdminOrderStatusUpdateInput{
		OrderID:            uuid.New(),
		Status:             domain.OrderCancelled,
		CancellationReason: "Out of stock",
	})
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPut, "/admin/orders/status", bytes.NewBuffer(body))
	c.Request.Header.Set("Content-Type", "application/json")

	ctl.AdminUpdateOrderStatus(c)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}
