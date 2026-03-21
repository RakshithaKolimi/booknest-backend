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
	"gorm.io/gorm"

	"booknest/internal/domain"
)

type mockReviewServiceController struct {
	listFunc   func(ctx context.Context, bookID uuid.UUID) (*domain.ReviewListResponse, error)
	upsertFunc func(ctx context.Context, bookID, userID uuid.UUID, input domain.ReviewInput) (*domain.Review, error)
}

func (m *mockReviewServiceController) ListBookReviews(ctx context.Context, bookID uuid.UUID) (*domain.ReviewListResponse, error) {
	if m.listFunc != nil {
		return m.listFunc(ctx, bookID)
	}
	return &domain.ReviewListResponse{}, nil
}

func (m *mockReviewServiceController) UpsertBookReview(ctx context.Context, bookID, userID uuid.UUID, input domain.ReviewInput) (*domain.Review, error) {
	if m.upsertFunc != nil {
		return m.upsertFunc(ctx, bookID, userID, input)
	}
	return nil, errors.New("not implemented")
}

func TestReviewControllerListBookReviews(t *testing.T) {
	gin.SetMode(gin.TestMode)
	bookID := uuid.New()
	ctl := NewReviewController(&mockReviewServiceController{
		listFunc: func(ctx context.Context, gotBookID uuid.UUID) (*domain.ReviewListResponse, error) {
			if gotBookID != bookID {
				t.Fatalf("unexpected book id: %s", gotBookID)
			}
			return &domain.ReviewListResponse{
				Items: []domain.Review{{ID: uuid.New(), BookID: bookID, Rating: 5}},
				Summary: domain.ReviewSummary{
					AverageRating: 5,
					TotalReviews:  1,
				},
			}, nil
		},
	}).(*reviewController)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{{Key: "id", Value: bookID.String()}}
	c.Request = httptest.NewRequest(http.MethodGet, "/books/"+bookID.String()+"/reviews", nil)

	ctl.ListBookReviews(c)
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestReviewControllerUpsertBookReview(t *testing.T) {
	gin.SetMode(gin.TestMode)
	bookID := uuid.New()
	userID := uuid.New()
	ctl := NewReviewController(&mockReviewServiceController{
		upsertFunc: func(ctx context.Context, gotBookID, gotUserID uuid.UUID, input domain.ReviewInput) (*domain.Review, error) {
			if gotBookID != bookID || gotUserID != userID {
				t.Fatalf("unexpected ids")
			}
			return &domain.Review{
				ID:      uuid.New(),
				BookID:  bookID,
				UserID:  userID,
				Rating:  input.Rating,
				Comment: input.Comment,
			}, nil
		},
	}).(*reviewController)

	payload, _ := json.Marshal(domain.ReviewInput{Rating: 5, Comment: "Loved it"})
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Params = gin.Params{{Key: "id", Value: bookID.String()}}
	c.Set("user_id", userID.String())
	c.Request = httptest.NewRequest(http.MethodPost, "/books/"+bookID.String()+"/reviews", bytes.NewBuffer(payload))
	c.Request.Header.Set("Content-Type", "application/json")

	ctl.UpsertBookReview(c)
	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", w.Code)
	}
}

func TestReviewControllerErrorCases(t *testing.T) {
	gin.SetMode(gin.TestMode)
	bookID := uuid.New()
	ctl := NewReviewController(&mockReviewServiceController{
		upsertFunc: func(ctx context.Context, bookID, userID uuid.UUID, input domain.ReviewInput) (*domain.Review, error) {
			return nil, gorm.ErrRecordNotFound
		},
	}).(*reviewController)

	t.Run("list invalid id", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = gin.Params{{Key: "id", Value: "bad-id"}}
		c.Request = httptest.NewRequest(http.MethodGet, "/books/bad-id/reviews", nil)

		ctl.ListBookReviews(c)
		if w.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d", w.Code)
		}
	})

	t.Run("upsert missing user", func(t *testing.T) {
		payload, _ := json.Marshal(domain.ReviewInput{Rating: 4, Comment: "Nice"})
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = gin.Params{{Key: "id", Value: bookID.String()}}
		c.Request = httptest.NewRequest(http.MethodPost, "/books/"+bookID.String()+"/reviews", bytes.NewBuffer(payload))
		c.Request.Header.Set("Content-Type", "application/json")

		ctl.UpsertBookReview(c)
		if w.Code != http.StatusUnauthorized {
			t.Fatalf("expected 401, got %d", w.Code)
		}
	})

	t.Run("upsert invalid id", func(t *testing.T) {
		payload, _ := json.Marshal(domain.ReviewInput{Rating: 4, Comment: "Nice"})
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = gin.Params{{Key: "id", Value: "bad-id"}}
		c.Set("user_id", uuid.New().String())
		c.Request = httptest.NewRequest(http.MethodPost, "/books/bad-id/reviews", bytes.NewBuffer(payload))
		c.Request.Header.Set("Content-Type", "application/json")

		ctl.UpsertBookReview(c)
		if w.Code != http.StatusBadRequest {
			t.Fatalf("expected 400, got %d", w.Code)
		}
	})

	t.Run("upsert missing book", func(t *testing.T) {
		payload, _ := json.Marshal(domain.ReviewInput{Rating: 4, Comment: "Nice"})
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = gin.Params{{Key: "id", Value: bookID.String()}}
		c.Set("user_id", uuid.New().String())
		c.Request = httptest.NewRequest(http.MethodPost, "/books/"+bookID.String()+"/reviews", bytes.NewBuffer(payload))
		c.Request.Header.Set("Content-Type", "application/json")

		ctl.UpsertBookReview(c)
		if w.Code != http.StatusNotFound {
			t.Fatalf("expected 404, got %d", w.Code)
		}
	})

	t.Run("upsert requires purchase", func(t *testing.T) {
		ctl := NewReviewController(&mockReviewServiceController{
			upsertFunc: func(ctx context.Context, bookID, userID uuid.UUID, input domain.ReviewInput) (*domain.Review, error) {
				return nil, domain.ErrReviewRequiresPurchase
			},
		}).(*reviewController)

		payload, _ := json.Marshal(domain.ReviewInput{Rating: 4, Comment: "Nice"})
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Params = gin.Params{{Key: "id", Value: bookID.String()}}
		c.Set("user_id", uuid.New().String())
		c.Request = httptest.NewRequest(http.MethodPost, "/books/"+bookID.String()+"/reviews", bytes.NewBuffer(payload))
		c.Request.Header.Set("Content-Type", "application/json")

		ctl.UpsertBookReview(c)
		if w.Code != http.StatusForbidden {
			t.Fatalf("expected 403, got %d", w.Code)
		}
	})
}
