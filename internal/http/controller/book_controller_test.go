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

type mockBookServiceController struct {
	createBookFunc      func(ctx context.Context, input domain.BookInput) (*domain.Book, error)
	getBookFunc         func(ctx context.Context, id uuid.UUID) (*domain.Book, error)
	listBooksFunc       func(ctx context.Context, limit, offset int) ([]domain.Book, error)
	filterByCriteriaFun func(ctx context.Context, filter domain.BookFilter, q domain.QueryOptions) (*domain.BookSearchResult, error)
	updateBookFunc      func(ctx context.Context, id uuid.UUID, input domain.BookInput) (*domain.Book, error)
	deleteBookFunc      func(ctx context.Context, id uuid.UUID) error
}

func (m *mockBookServiceController) CreateBook(ctx context.Context, input domain.BookInput) (*domain.Book, error) {
	if m.createBookFunc != nil {
		return m.createBookFunc(ctx, input)
	}
	return nil, errors.New("not implemented")
}
func (m *mockBookServiceController) GetBook(ctx context.Context, id uuid.UUID) (*domain.Book, error) {
	if m.getBookFunc != nil {
		return m.getBookFunc(ctx, id)
	}
	return nil, errors.New("not implemented")
}
func (m *mockBookServiceController) ListBooks(ctx context.Context, limit, offset int) ([]domain.Book, error) {
	if m.listBooksFunc != nil {
		return m.listBooksFunc(ctx, limit, offset)
	}
	return []domain.Book{}, nil
}
func (m *mockBookServiceController) FilterByCriteria(ctx context.Context, filter domain.BookFilter, q domain.QueryOptions) (*domain.BookSearchResult, error) {
	if m.filterByCriteriaFun != nil {
		return m.filterByCriteriaFun(ctx, filter, q)
	}
	return &domain.BookSearchResult{}, nil
}
func (m *mockBookServiceController) UpdateBook(ctx context.Context, id uuid.UUID, input domain.BookInput) (*domain.Book, error) {
	if m.updateBookFunc != nil {
		return m.updateBookFunc(ctx, id, input)
	}
	return nil, errors.New("not implemented")
}
func (m *mockBookServiceController) DeleteBook(ctx context.Context, id uuid.UUID) error {
	if m.deleteBookFunc != nil {
		return m.deleteBookFunc(ctx, id)
	}
	return nil
}

func TestBookControllerGetAndList(t *testing.T) {
	gin.SetMode(gin.TestMode)
	id := uuid.New()
	svc := &mockBookServiceController{
		getBookFunc: func(ctx context.Context, gotID uuid.UUID) (*domain.Book, error) {
			if gotID != id {
				t.Fatalf("unexpected id: %s", gotID)
			}
			return &domain.Book{ID: id, Name: "Book"}, nil
		},
			listBooksFunc: func(ctx context.Context, limit, offset int) ([]domain.Book, error) {
				if limit != 500 || offset != 0 {
					t.Fatalf("expected defaults 500/0, got %d/%d", limit, offset)
				}
				return []domain.Book{{ID: id, Name: "Book"}}, nil
			},
	}
	ctl := NewBookController(svc).(*bookController)

	gw := httptest.NewRecorder()
	gc, _ := gin.CreateTestContext(gw)
	gc.Params = gin.Params{{Key: "id", Value: id.String()}}
	gc.Request = httptest.NewRequest(http.MethodGet, "/books/"+id.String(), nil)
	ctl.getBook(gc)
	if gw.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", gw.Code)
	}

	lw := httptest.NewRecorder()
	lc, _ := gin.CreateTestContext(lw)
	lc.Request = httptest.NewRequest(http.MethodGet, "/books", nil)
	ctl.listBooks(lc)
	if lw.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", lw.Code)
	}
}

func TestBookControllerFilterBooks(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := &mockBookServiceController{
		filterByCriteriaFun: func(ctx context.Context, filter domain.BookFilter, q domain.QueryOptions) (*domain.BookSearchResult, error) {
			if filter.Search == nil || *filter.Search != "harry" {
				t.Fatalf("expected search query to be set")
			}
			if q.Limit != 5 || q.Offset != 2 {
				t.Fatalf("unexpected pagination: %+v", q)
			}
			return &domain.BookSearchResult{Items: []domain.Book{}, Total: 0, Limit: q.Limit, Offset: q.Offset}, nil
		},
	}
	ctl := NewBookController(svc).(*bookController)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/books/filter?search=harry&limit=5&offset=2", bytes.NewBufferString(`{}`))
	c.Request.Header.Set("Content-Type", "application/json")

	ctl.filterBooks(c)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestBookControllerCreateValidation(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ctl := NewBookController(&mockBookServiceController{}).(*bookController)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	payload, _ := json.Marshal(map[string]any{"name": "Only Name"})
	c.Request = httptest.NewRequest(http.MethodPost, "/books", bytes.NewBuffer(payload))
	c.Request.Header.Set("Content-Type", "application/json")

	ctl.createBook(c)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}
