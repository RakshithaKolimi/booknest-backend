package controller

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"booknest/internal/domain"
)

func TestPublisherControllerList_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)

	svc := &MockPublisherService{
		ListFunc: func(ctx context.Context, limit, offset int, search string) ([]domain.Publisher, error) {
			if limit != 5 || offset != 2 {
				t.Fatalf("unexpected pagination %d/%d", limit, offset)
			}
			if search != "Acme" {
				t.Fatalf("search should be trimmed, got %q", search)
			}
			return []domain.Publisher{{ID: uuid.New(), LegalName: "Acme"}}, nil
		},
	}
	ctl := NewPublisherController(svc).(*publisherController)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/publishers?limit=5&offset=2&search=%20%20Acme%20", nil)

	ctl.List(c)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestPublisherControllerGetByID_InvalidID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ctl := NewPublisherController(&MockPublisherService{}).(*publisherController)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{{Key: "id", Value: "bad-id"}}
	c.Request = httptest.NewRequest(http.MethodGet, "/publishers/bad-id", nil)

	ctl.GetByID(c)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestPublisherControllerCreate_SanitizesInput(t *testing.T) {
	gin.SetMode(gin.TestMode)

	svc := &MockPublisherService{
		CreateFunc: func(ctx context.Context, in domain.PublisherInput) (*domain.Publisher, error) {
			if in.LegalName != "Legal Name" {
				t.Fatalf("expected normalized legal name, got %q", in.LegalName)
			}
			if in.Address != "One Main St" {
				t.Fatalf("expected normalized address, got %q", in.Address)
			}
			return &domain.Publisher{ID: uuid.New(), LegalName: in.LegalName}, nil
		},
	}
	ctl := NewPublisherController(svc).(*publisherController)

	payload := domain.PublisherInput{
		LegalName:   "  Ｌｅｇａｌ   Name  ",
		TradingName: "  Trade  Name ",
		Email:       "test@mail.com",
		Mobile:      "+911234567890",
		Address:     " One\tMain\nSt ",
		City:        " City ",
		State:       " State ",
		Country:     " Country ",
		Zipcode:     " 123456 ",
	}
	body, _ := json.Marshal(payload)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/publishers", bytes.NewBuffer(body))
	c.Request.Header.Set("Content-Type", "application/json")

	ctl.Create(c)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d body=%s", w.Code, w.Body.String())
	}
}

func TestPublisherControllerCreate_RejectsExcessiveLength(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ctl := NewPublisherController(&MockPublisherService{}).(*publisherController)

	payload := domain.PublisherInput{
		LegalName:   strings.Repeat("x", defaultMaxInputLength+1),
		TradingName: "Trading Name",
		Email:       "test@mail.com",
		Mobile:      "+911234567890",
		Address:     "123 Main St",
		City:        "City",
		State:       "State",
		Country:     "Country",
		Zipcode:     "123456",
	}
	body, _ := json.Marshal(payload)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/publishers", bytes.NewBuffer(body))
	c.Request.Header.Set("Content-Type", "application/json")

	ctl.Create(c)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestPublisherControllerSetActive_AndDelete(t *testing.T) {
	gin.SetMode(gin.TestMode)

	id := uuid.New()
	svc := &MockPublisherService{
		SetActiveFunc: func(ctx context.Context, got uuid.UUID, active bool) error {
			if got != id || !active {
				t.Fatalf("unexpected set-active args")
			}
			return nil
		},
		DeleteFunc: func(ctx context.Context, got uuid.UUID) error {
			if got != id {
				t.Fatalf("unexpected delete id %s", got)
			}
			return nil
		},
	}
	ctl := NewPublisherController(svc).(*publisherController)

	setBody, _ := json.Marshal(map[string]bool{"active": true})
	sw := httptest.NewRecorder()
	sc, _ := gin.CreateTestContext(sw)
	sc.Params = gin.Params{{Key: "id", Value: id.String()}}
	sc.Request = httptest.NewRequest(http.MethodPatch, "/publishers/"+id.String()+"/status", bytes.NewBuffer(setBody))
	sc.Request.Header.Set("Content-Type", "application/json")
	ctl.SetActive(sc)
	if sw.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", sw.Code)
	}

	dw := httptest.NewRecorder()
	dc, _ := gin.CreateTestContext(dw)
	dc.Params = gin.Params{{Key: "id", Value: id.String()}}
	dc.Request = httptest.NewRequest(http.MethodDelete, "/publishers/"+id.String(), nil)
	ctl.Delete(dc)
	if dw.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", dw.Code)
	}
}

func TestPublisherControllerErrorPaths(t *testing.T) {
	gin.SetMode(gin.TestMode)

	id := uuid.New()
	svc := &MockPublisherService{
		FindFunc: func(ctx context.Context, id uuid.UUID) (*domain.Publisher, error) {
			return nil, errors.New("not found")
		},
		UpdateFunc: func(ctx context.Context, id uuid.UUID, in domain.PublisherInput) (*domain.Publisher, error) {
			return nil, errors.New("not found")
		},
	}
	ctl := NewPublisherController(svc).(*publisherController)

	gw := httptest.NewRecorder()
	gc, _ := gin.CreateTestContext(gw)
	gc.Params = gin.Params{{Key: "id", Value: id.String()}}
	gc.Request = httptest.NewRequest(http.MethodGet, "/publishers/"+id.String(), nil)
	ctl.GetByID(gc)
	if gw.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", gw.Code)
	}

	updatePayload := domain.PublisherInput{
		LegalName:   "Legal",
		TradingName: "Trading",
		Email:       "test@mail.com",
		Mobile:      "+911234567890",
		Address:     "123 Main St",
		City:        "City",
		State:       "State",
		Country:     "Country",
		Zipcode:     "123456",
	}
	body, _ := json.Marshal(updatePayload)
	uw := httptest.NewRecorder()
	uc, _ := gin.CreateTestContext(uw)
	uc.Params = gin.Params{{Key: "id", Value: id.String()}}
	uc.Request = httptest.NewRequest(http.MethodPut, "/publishers/"+id.String(), bytes.NewBuffer(body))
	uc.Request.Header.Set("Content-Type", "application/json")
	ctl.Update(uc)
	if uw.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", uw.Code)
	}
}

func TestPublisherControllerAdditionalErrorPaths(t *testing.T) {
	gin.SetMode(gin.TestMode)

	id := uuid.New()
	ctl := NewPublisherController(&MockPublisherService{
		ListFunc: func(ctx context.Context, limit, offset int, search string) ([]domain.Publisher, error) {
			return nil, errors.New("boom")
		},
		CreateFunc: func(ctx context.Context, in domain.PublisherInput) (*domain.Publisher, error) {
			return nil, errors.New("boom")
		},
		SetActiveFunc: func(ctx context.Context, id uuid.UUID, active bool) error { return errors.New("boom") },
		DeleteFunc:    func(ctx context.Context, id uuid.UUID) error { return errors.New("boom") },
	}).(*publisherController)

	tests := []struct {
		name   string
		method func(*gin.Context)
		setup  func(*gin.Context)
		body   string
		code   int
	}{
		{name: "list error", method: ctl.List, code: http.StatusInternalServerError},
		{name: "create error", method: ctl.Create, body: `{"legal_name":"Legal","trading_name":"Trading","email":"test@mail.com","mobile":"+911234567890","address":"Addr","city":"City","state":"State","country":"Country","zipcode":"123456"}`, code: http.StatusInternalServerError},
		{name: "update invalid id", method: ctl.Update, setup: func(c *gin.Context) { c.Params = gin.Params{{Key: "id", Value: "bad-id"}} }, body: `{}`, code: http.StatusBadRequest},
		{name: "set active invalid id", method: ctl.SetActive, setup: func(c *gin.Context) { c.Params = gin.Params{{Key: "id", Value: "bad-id"}} }, body: `{"active":true}`, code: http.StatusBadRequest},
		{name: "set active error", method: ctl.SetActive, setup: func(c *gin.Context) { c.Params = gin.Params{{Key: "id", Value: id.String()}} }, body: `{"active":true}`, code: http.StatusInternalServerError},
		{name: "delete invalid id", method: ctl.Delete, setup: func(c *gin.Context) { c.Params = gin.Params{{Key: "id", Value: "bad-id"}} }, code: http.StatusBadRequest},
		{name: "delete error", method: ctl.Delete, setup: func(c *gin.Context) { c.Params = gin.Params{{Key: "id", Value: id.String()}} }, code: http.StatusInternalServerError},
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
