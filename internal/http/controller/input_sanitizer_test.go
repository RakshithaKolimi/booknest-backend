package controller

import (
	"strings"
	"testing"
)

type sanitizeNested struct {
	Note string `json:"note"`
}

type sanitizeSample struct {
	Name     string           `json:"name"`
	Email    string           `json:"email"`
	Password string           `json:"password"`
	Comment  *string          `json:"comment,omitempty"`
	Meta     []sanitizeNested `json:"meta,omitempty"`
	RawToken string           `json:"raw_token"`
}

func TestSanitizeInput_TrimAndNormalize(t *testing.T) {
	comment := "  hello\x00\tworld  "
	input := sanitizeSample{
		Name:     "  Ｊｏｈｎ   Doe  ",
		Email:    "  test@example.com  ",
		Password: "  pass  word  ",
		Comment:  &comment,
		Meta:     []sanitizeNested{{Note: "  line\u00A0space  "}},
	}

	if err := sanitizeInput(&input); err != nil {
		t.Fatalf("sanitizeInput returned error: %v", err)
	}

	if input.Name != "John Doe" {
		t.Fatalf("unexpected normalized name: %q", input.Name)
	}
	if input.Email != "test@example.com" {
		t.Fatalf("unexpected email: %q", input.Email)
	}
	if input.Password != "pass  word" {
		t.Fatalf("password should preserve internal whitespace: %q", input.Password)
	}
	if input.Comment == nil || *input.Comment != "hello world" {
		t.Fatalf("unexpected comment normalization: %v", input.Comment)
	}
	if got := input.Meta[0].Note; got != "line space" {
		t.Fatalf("unexpected nested note normalization: %q", got)
	}
}

func TestSanitizeInput_RejectsExcessiveLength(t *testing.T) {
	input := sanitizeSample{
		Email: strings.Repeat("a", emailMaxInputLength+1) + "@x.com",
	}

	if err := sanitizeInput(&input); err == nil {
		t.Fatal("expected error for excessive input length")
	}
}

func TestSanitizeInput_RejectsNilPointer(t *testing.T) {
	if err := sanitizeInput(nil); err == nil {
		t.Fatal("expected error for nil input")
	}
}
