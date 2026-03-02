package controller

import (
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"booknest/internal/domain"
)

func TestGetUserID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	if _, err := getUserID(c); err == nil {
		t.Fatal("expected error for missing user_id")
	}

	c.Set("user_id", "bad-id")
	if _, err := getUserID(c); err == nil {
		t.Fatal("expected parse error for bad uuid")
	}

	id := uuid.New()
	c.Set("user_id", id.String())
	got, err := getUserID(c)
	if err != nil || got != id {
		t.Fatalf("expected parsed uuid, got %v err=%v", got, err)
	}

	c.Set("user_id", id)
	got, err = getUserID(c)
	if err != nil || got != id {
		t.Fatalf("expected uuid passthrough, got %v err=%v", got, err)
	}

	c.Set("user_id", 123)
	if _, err := getUserID(c); err == nil {
		t.Fatal("expected error for invalid type")
	}
}

func TestGetUserRole(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	if _, err := getUserRole(c); err == nil {
		t.Fatal("expected error for missing user_role")
	}

	c.Set("user_role", "ADMIN")
	role, err := getUserRole(c)
	if err != nil || role != domain.UserRole("ADMIN") {
		t.Fatalf("unexpected role %q err=%v", role, err)
	}

	c.Set("user_role", domain.UserRole("USER"))
	role, err = getUserRole(c)
	if err != nil || role != domain.UserRole("USER") {
		t.Fatalf("unexpected role %q err=%v", role, err)
	}

	c.Set("user_role", 100)
	if _, err := getUserRole(c); err == nil {
		t.Fatal("expected error for invalid user_role type")
	}
}
