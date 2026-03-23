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

type mockAuthorService struct {
	findByIDFunc func(ctx context.Context, id uuid.UUID) (*domain.Author, error)
	listFunc     func(ctx context.Context, limit, offset int, search string) ([]domain.Author, error)
	createFunc   func(ctx context.Context, input domain.AuthorInput) (*domain.Author, error)
	updateFunc   func(ctx context.Context, id uuid.UUID, input domain.AuthorInput) (*domain.Author, error)
	deleteFunc   func(ctx context.Context, id uuid.UUID) error
}

func (m *mockAuthorService) FindByID(ctx context.Context, id uuid.UUID) (*domain.Author, error) {
	if m.findByIDFunc != nil {
		return m.findByIDFunc(ctx, id)
	}
	return nil, errors.New("not implemented")
}

func (m *mockAuthorService) List(ctx context.Context, limit, offset int, search string) ([]domain.Author, error) {
	if m.listFunc != nil {
		return m.listFunc(ctx, limit, offset, search)
	}
	return []domain.Author{}, nil
}

func (m *mockAuthorService) Create(ctx context.Context, input domain.AuthorInput) (*domain.Author, error) {
	if m.createFunc != nil {
		return m.createFunc(ctx, input)
	}
	return nil, errors.New("not implemented")
}

func (m *mockAuthorService) Update(ctx context.Context, id uuid.UUID, input domain.AuthorInput) (*domain.Author, error) {
	if m.updateFunc != nil {
		return m.updateFunc(ctx, id, input)
	}
	return nil, errors.New("not implemented")
}

func (m *mockAuthorService) Delete(ctx context.Context, id uuid.UUID) error {
	if m.deleteFunc != nil {
		return m.deleteFunc(ctx, id)
	}
	return nil
}

func TestAuthorControllerCreate(t *testing.T) {
	gin.SetMode(gin.TestMode)

	svc := &mockAuthorService{createFunc: func(ctx context.Context, input domain.AuthorInput) (*domain.Author, error) {
		if input.Name != "Author Name" {
			t.Fatalf("unexpected input: %+v", input)
		}
		return &domain.Author{ID: uuid.New(), Name: input.Name}, nil
	}}
	ctl := NewAuthorController(svc).(*authorController)

	body, _ := json.Marshal(domain.AuthorInput{Name: "Author Name"})
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/authors", bytes.NewBuffer(body))
	c.Request.Header.Set("Content-Type", "application/json")

	ctl.Create(c)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", w.Code)
	}
}

func TestAuthorControllerListAndGetByID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	id := uuid.New()
	svc := &mockAuthorService{
		listFunc: func(ctx context.Context, limit, offset int, search string) ([]domain.Author, error) {
			if limit != 7 || offset != 3 || search != "tol" {
				t.Fatalf("unexpected list params: %d/%d %q", limit, offset, search)
			}
			return []domain.Author{{ID: id, Name: "A"}}, nil
		},
		findByIDFunc: func(ctx context.Context, gotID uuid.UUID) (*domain.Author, error) {
			if gotID != id {
				t.Fatalf("unexpected id: %s", gotID)
			}
			return &domain.Author{ID: id, Name: "A"}, nil
		},
	}
	ctl := NewAuthorController(svc).(*authorController)

	lw := httptest.NewRecorder()
	lc, _ := gin.CreateTestContext(lw)
	lc.Request = httptest.NewRequest(http.MethodGet, "/authors?limit=7&offset=3&search=tol", nil)
	ctl.List(lc)
	if lw.Code != http.StatusOK {
		t.Fatalf("expected 200 from list, got %d", lw.Code)
	}

	gw := httptest.NewRecorder()
	gc, _ := gin.CreateTestContext(gw)
	gc.Params = gin.Params{{Key: "id", Value: id.String()}}
	gc.Request = httptest.NewRequest(http.MethodGet, "/authors/"+id.String(), nil)
	ctl.GetByID(gc)
	if gw.Code != http.StatusOK {
		t.Fatalf("expected 200 from getByID, got %d", gw.Code)
	}
}

func TestAuthorControllerDeleteInvalidID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	ctl := NewAuthorController(&mockAuthorService{}).(*authorController)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{{Key: "id", Value: "bad-id"}}
	c.Request = httptest.NewRequest(http.MethodDelete, "/authors/bad-id", nil)

	ctl.Delete(c)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", w.Code)
	}
}

func TestAuthorControllerUpdateAndDelete(t *testing.T) {
	gin.SetMode(gin.TestMode)
	id := uuid.New()
	updated := false
	deleted := false
	svc := &mockAuthorService{
		updateFunc: func(ctx context.Context, gotID uuid.UUID, input domain.AuthorInput) (*domain.Author, error) {
			updated = true
			if gotID != id || input.Name != "Updated Author" {
				t.Fatalf("unexpected update input")
			}
			return &domain.Author{ID: id, Name: input.Name}, nil
		},
		deleteFunc: func(ctx context.Context, gotID uuid.UUID) error {
			deleted = true
			if gotID != id {
				t.Fatalf("unexpected delete id")
			}
			return nil
		},
	}
	ctl := NewAuthorController(svc).(*authorController)

	body, _ := json.Marshal(domain.AuthorInput{Name: "Updated Author"})
	uw := httptest.NewRecorder()
	uc, _ := gin.CreateTestContext(uw)
	uc.Params = gin.Params{{Key: "id", Value: id.String()}}
	uc.Request = httptest.NewRequest(http.MethodPut, "/authors/"+id.String(), bytes.NewBuffer(body))
	uc.Request.Header.Set("Content-Type", "application/json")
	ctl.Update(uc)
	if uw.Code != http.StatusOK || !updated {
		t.Fatalf("expected update success, code=%d updated=%v", uw.Code, updated)
	}

	dw := httptest.NewRecorder()
	dc, _ := gin.CreateTestContext(dw)
	dc.Params = gin.Params{{Key: "id", Value: id.String()}}
	dc.Request = httptest.NewRequest(http.MethodDelete, "/authors/"+id.String(), nil)
	ctl.Delete(dc)
	if dw.Code != http.StatusOK || !deleted {
		t.Fatalf("expected delete success, code=%d deleted=%v", dw.Code, deleted)
	}
}

func TestAuthorControllerErrorPaths(t *testing.T) {
	gin.SetMode(gin.TestMode)
	id := uuid.New()
	ctl := NewAuthorController(&mockAuthorService{
		createFunc: func(ctx context.Context, input domain.AuthorInput) (*domain.Author, error) {
			return nil, errors.New("create failed")
		},
		findByIDFunc: func(ctx context.Context, gotID uuid.UUID) (*domain.Author, error) {
			return nil, errors.New("not found")
		},
		updateFunc: func(ctx context.Context, gotID uuid.UUID, input domain.AuthorInput) (*domain.Author, error) {
			return nil, errors.New("update failed")
		},
		deleteFunc: func(ctx context.Context, gotID uuid.UUID) error { return errors.New("delete failed") },
	}).(*authorController)

	cases := []struct {
		name   string
		method func(*gin.Context)
		setup  func(*gin.Context)
		body   string
		code   int
	}{
		{name: "create error", method: ctl.Create, body: `{"name":"Author Name"}`, code: http.StatusBadRequest},
		{name: "get not found", method: ctl.GetByID, setup: func(c *gin.Context) { c.Params = gin.Params{{Key: "id", Value: id.String()}} }, code: http.StatusNotFound},
		{name: "update error", method: ctl.Update, setup: func(c *gin.Context) { c.Params = gin.Params{{Key: "id", Value: id.String()}} }, body: `{"name":"Updated Author"}`, code: http.StatusBadRequest},
		{name: "delete error", method: ctl.Delete, setup: func(c *gin.Context) { c.Params = gin.Params{{Key: "id", Value: id.String()}} }, code: http.StatusInternalServerError},
	}

	for _, tc := range cases {
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
