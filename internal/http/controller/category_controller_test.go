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

type mockCategoryService struct {
	findByIDFunc func(ctx context.Context, id uuid.UUID) (*domain.Category, error)
	listFunc     func(ctx context.Context, limit, offset int) ([]domain.Category, error)
	createFunc   func(ctx context.Context, input domain.CategoryInput) (*domain.Category, error)
	updateFunc   func(ctx context.Context, id uuid.UUID, input domain.CategoryInput) (*domain.Category, error)
	deleteFunc   func(ctx context.Context, id uuid.UUID) error
}

func (m *mockCategoryService) FindByID(ctx context.Context, id uuid.UUID) (*domain.Category, error) {
	if m.findByIDFunc != nil {
		return m.findByIDFunc(ctx, id)
	}
	return nil, errors.New("not implemented")
}
func (m *mockCategoryService) List(ctx context.Context, limit, offset int) ([]domain.Category, error) {
	if m.listFunc != nil {
		return m.listFunc(ctx, limit, offset)
	}
	return []domain.Category{}, nil
}
func (m *mockCategoryService) Create(ctx context.Context, input domain.CategoryInput) (*domain.Category, error) {
	if m.createFunc != nil {
		return m.createFunc(ctx, input)
	}
	return nil, errors.New("not implemented")
}
func (m *mockCategoryService) Update(ctx context.Context, id uuid.UUID, input domain.CategoryInput) (*domain.Category, error) {
	if m.updateFunc != nil {
		return m.updateFunc(ctx, id, input)
	}
	return nil, errors.New("not implemented")
}
func (m *mockCategoryService) Delete(ctx context.Context, id uuid.UUID) error {
	if m.deleteFunc != nil {
		return m.deleteFunc(ctx, id)
	}
	return nil
}

func TestCategoryControllerCreateAndGetByID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	id := uuid.New()
	svc := &mockCategoryService{
		createFunc: func(ctx context.Context, input domain.CategoryInput) (*domain.Category, error) {
			return &domain.Category{ID: id, Name: input.Name}, nil
		},
		findByIDFunc: func(ctx context.Context, gotID uuid.UUID) (*domain.Category, error) {
			if gotID != id {
				t.Fatalf("unexpected id: %s", gotID)
			}
			return &domain.Category{ID: id, Name: "Fiction"}, nil
		},
	}
	ctl := NewCategoryController(svc).(*categoryController)

	body, _ := json.Marshal(domain.CategoryInput{Name: "Fiction"})
	cw := httptest.NewRecorder()
	cc, _ := gin.CreateTestContext(cw)
	cc.Request = httptest.NewRequest(http.MethodPost, "/categories", bytes.NewBuffer(body))
	cc.Request.Header.Set("Content-Type", "application/json")
	ctl.Create(cc)
	if cw.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", cw.Code)
	}

	gw := httptest.NewRecorder()
	gc, _ := gin.CreateTestContext(gw)
	gc.Params = gin.Params{{Key: "id", Value: id.String()}}
	gc.Request = httptest.NewRequest(http.MethodGet, "/categories/"+id.String(), nil)
	ctl.GetByID(gc)
	if gw.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", gw.Code)
	}
}

func TestCategoryControllerListDefaults(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := &mockCategoryService{listFunc: func(ctx context.Context, limit, offset int) ([]domain.Category, error) {
		if limit != 500 || offset != 0 {
			t.Fatalf("expected defaults 20/0, got %d/%d", limit, offset)
		}
		return []domain.Category{}, nil
	}}
	ctl := NewCategoryController(svc).(*categoryController)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/categories", nil)

	ctl.List(c)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestCategoryControllerUpdateBadID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ctl := NewCategoryController(&mockCategoryService{}).(*categoryController)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{{Key: "id", Value: "bad-id"}}
	c.Request = httptest.NewRequest(http.MethodPut, "/categories/bad-id", bytes.NewBufferString(`{"name":"X"}`))
	c.Request.Header.Set("Content-Type", "application/json")

	ctl.Update(c)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}
