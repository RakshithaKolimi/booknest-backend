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
	queryBooksFunc      func(ctx context.Context, filter domain.BookFilter, q domain.QueryOptions) (*domain.BookSearchResult, error)
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
func (m *mockBookServiceController) QueryBooks(ctx context.Context, filter domain.BookFilter, q domain.QueryOptions) (*domain.BookSearchResult, error) {
	if m.queryBooksFunc != nil {
		return m.queryBooksFunc(ctx, filter, q)
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

func TestBookControllerQueryBooks(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := &mockBookServiceController{
		queryBooksFunc: func(ctx context.Context, filter domain.BookFilter, q domain.QueryOptions) (*domain.BookSearchResult, error) {
			if q.Limit != 12 {
				t.Fatalf("expected default limit 12, got %d", q.Limit)
			}
			if q.Cursor == nil || *q.Cursor != "abc" {
				t.Fatalf("expected cursor to be set")
			}
			return &domain.BookSearchResult{
				Items:      []domain.Book{},
				Total:      0,
				Limit:      q.Limit,
				Offset:     q.Offset,
				NextCursor: nil,
				HasMore:    false,
			}, nil
		},
	}
	ctl := NewBookController(svc).(*bookController)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(
		http.MethodGet,
		"/books/search?book_name=Harry&author_name=Rowling&cursor=abc",
		nil,
	)

	ctl.queryBooks(c)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestBookControllerCreateUpdateDeleteAndErrors(t *testing.T) {
	gin.SetMode(gin.TestMode)
	id := uuid.New()
	svc := &mockBookServiceController{
		createBookFunc: func(ctx context.Context, input domain.BookInput) (*domain.Book, error) {
			return &domain.Book{ID: id, Name: input.Name}, nil
		},
		updateBookFunc: func(ctx context.Context, gotID uuid.UUID, input domain.BookInput) (*domain.Book, error) {
			return &domain.Book{ID: gotID, Name: input.Name}, nil
		},
		deleteBookFunc: func(ctx context.Context, gotID uuid.UUID) error { return nil },
	}
	ctl := NewBookController(svc).(*bookController)

	payload := `{"name":"Book","author_name":"Author","publisher_id":"` + uuid.New().String() + `","price":10}`

	cw := httptest.NewRecorder()
	cc, _ := gin.CreateTestContext(cw)
	cc.Request = httptest.NewRequest(http.MethodPost, "/books", bytes.NewBufferString(payload))
	cc.Request.Header.Set("Content-Type", "application/json")
	ctl.createBook(cc)
	if cw.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", cw.Code)
	}

	uw := httptest.NewRecorder()
	uc, _ := gin.CreateTestContext(uw)
	uc.Params = gin.Params{{Key: "id", Value: id.String()}}
	uc.Request = httptest.NewRequest(http.MethodPut, "/books/"+id.String(), bytes.NewBufferString(payload))
	uc.Request.Header.Set("Content-Type", "application/json")
	ctl.updateBook(uc)
	if uw.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", uw.Code)
	}

	dw := httptest.NewRecorder()
	dc, _ := gin.CreateTestContext(dw)
	dc.Params = gin.Params{{Key: "id", Value: id.String()}}
	dc.Request = httptest.NewRequest(http.MethodDelete, "/books/"+id.String(), nil)
	ctl.deleteBook(dc)
	if dw.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", dw.Code)
	}
}

func TestBookControllerDirectErrorPaths(t *testing.T) {
	gin.SetMode(gin.TestMode)
	id := uuid.New()
	svc := &mockBookServiceController{
		createBookFunc: func(ctx context.Context, input domain.BookInput) (*domain.Book, error) {
			return nil, errors.New("boom")
		},
		getBookFunc:   func(ctx context.Context, gotID uuid.UUID) (*domain.Book, error) { return nil, errors.New("boom") },
		listBooksFunc: func(ctx context.Context, limit, offset int) ([]domain.Book, error) { return nil, errors.New("boom") },
		filterByCriteriaFun: func(ctx context.Context, filter domain.BookFilter, q domain.QueryOptions) (*domain.BookSearchResult, error) {
			return nil, errors.New("boom")
		},
		queryBooksFunc: func(ctx context.Context, filter domain.BookFilter, q domain.QueryOptions) (*domain.BookSearchResult, error) {
			return nil, errors.New("boom")
		},
		updateBookFunc: func(ctx context.Context, gotID uuid.UUID, input domain.BookInput) (*domain.Book, error) {
			return nil, errors.New("boom")
		},
		deleteBookFunc: func(ctx context.Context, gotID uuid.UUID) error { return errors.New("boom") },
	}
	ctl := NewBookController(svc).(*bookController)

	tests := []struct {
		name   string
		method func(*gin.Context)
		setup  func(*gin.Context)
		body   string
		url    string
		code   int
	}{
		{name: "get invalid id", method: ctl.getBook, setup: func(c *gin.Context) { c.Params = gin.Params{{Key: "id", Value: "bad"}} }, code: http.StatusBadRequest},
		{name: "get not found", method: ctl.getBook, setup: func(c *gin.Context) { c.Params = gin.Params{{Key: "id", Value: id.String()}} }, code: http.StatusNotFound},
		{name: "list error", method: ctl.listBooks, code: http.StatusInternalServerError},
		{name: "query invalid limit", method: ctl.queryBooks, url: "/books/search?limit=bad", code: http.StatusBadRequest},
		{name: "query invalid offset", method: ctl.queryBooks, url: "/books/search?offset=bad", code: http.StatusBadRequest},
		{name: "query error", method: ctl.queryBooks, url: "/books/search", code: http.StatusInternalServerError},
		{name: "filter error", method: ctl.filterBooks, body: `{}`, code: http.StatusInternalServerError},
		{name: "update invalid id", method: ctl.updateBook, setup: func(c *gin.Context) { c.Params = gin.Params{{Key: "id", Value: "bad"}} }, code: http.StatusBadRequest},
		{name: "update error", method: ctl.updateBook, setup: func(c *gin.Context) { c.Params = gin.Params{{Key: "id", Value: id.String()}} }, body: `{"name":"Book","author_name":"Author","publisher_id":"` + uuid.New().String() + `"}`, code: http.StatusInternalServerError},
		{name: "delete invalid id", method: ctl.deleteBook, setup: func(c *gin.Context) { c.Params = gin.Params{{Key: "id", Value: "bad"}} }, code: http.StatusBadRequest},
		{name: "delete error", method: ctl.deleteBook, setup: func(c *gin.Context) { c.Params = gin.Params{{Key: "id", Value: id.String()}} }, code: http.StatusInternalServerError},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			if tc.setup != nil {
				tc.setup(c)
			}
			url := tc.url
			if url == "" {
				url = "/test"
			}
			c.Request = httptest.NewRequest(http.MethodPost, url, bytes.NewBufferString(tc.body))
			c.Request.Header.Set("Content-Type", "application/json")
			tc.method(c)
			if w.Code != tc.code {
				t.Fatalf("expected %d, got %d", tc.code, w.Code)
			}
		})
	}
}

func TestBookControllerQueryBooksClampLimit(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ctl := NewBookController(&mockBookServiceController{
		queryBooksFunc: func(ctx context.Context, filter domain.BookFilter, q domain.QueryOptions) (*domain.BookSearchResult, error) {
			if q.Limit != 100 {
				t.Fatalf("expected limit to clamp to 100, got %d", q.Limit)
			}
			return &domain.BookSearchResult{Limit: q.Limit}, nil
		},
	}).(*bookController)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/books/search?limit=500", nil)

	ctl.queryBooks(c)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}
