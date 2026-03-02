package controller

import (
	"fmt"
	"reflect"
	"strings"
	"unicode"
	"unicode/utf8"

	"golang.org/x/text/unicode/norm"
)

const (
	defaultMaxInputLength  = 2048
	passwordMaxInputLength = 256
	tokenMaxInputLength    = 1024
	emailMaxInputLength    = 254
	mobileMaxInputLength   = 20
	otpMaxInputLength      = 12
)

func sanitizeInput(v any) error {
	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Ptr || rv.IsNil() {
		return fmt.Errorf("input must be a non-nil pointer")
	}
	return sanitizeValue(rv.Elem(), "")
}

func sanitizeValue(v reflect.Value, fieldPath string) error {
	switch v.Kind() {
	case reflect.Struct:
		t := v.Type()
		for i := 0; i < v.NumField(); i++ {
			structField := t.Field(i)
			if structField.PkgPath != "" { // unexported
				continue
			}
			childPath := joinFieldPath(fieldPath, structField)
			if err := sanitizeValue(v.Field(i), childPath); err != nil {
				return err
			}
		}
	case reflect.Ptr:
		if v.IsNil() {
			return nil
		}
		return sanitizeValue(v.Elem(), fieldPath)
	case reflect.Slice, reflect.Array:
		for i := 0; i < v.Len(); i++ {
			if err := sanitizeValue(v.Index(i), fieldPath); err != nil {
				return err
			}
		}
	case reflect.String:
		normalized := normalizeInputString(v.String(), isPasswordField(fieldPath))
		if utf8.RuneCountInString(normalized) > maxLengthForField(fieldPath) {
			return fmt.Errorf("%s exceeds maximum length", fieldPath)
		}
		if v.CanSet() {
			v.SetString(normalized)
		}
	}

	return nil
}

func normalizeInputString(value string, preserveInternalWhitespace bool) string {
	value = norm.NFKC.String(value)
	value = strings.Map(func(r rune) rune {
		if r == utf8.RuneError || unicode.Is(unicode.Cf, r) {
			return -1
		}
		if r == '\t' || r == '\n' || r == '\r' {
			return ' '
		}
		if unicode.IsControl(r) {
			return -1
		}
		if r == '\u00A0' {
			return ' '
		}
		return r
	}, value)

	value = strings.TrimSpace(value)
	if preserveInternalWhitespace {
		return value
	}

	return strings.Join(strings.Fields(value), " ")
}

func maxLengthForField(fieldPath string) int {
	lower := strings.ToLower(fieldPath)
	switch {
	case strings.Contains(lower, "password"):
		return passwordMaxInputLength
	case strings.Contains(lower, "token"):
		return tokenMaxInputLength
	case strings.Contains(lower, "email"):
		return emailMaxInputLength
	case strings.Contains(lower, "mobile"):
		return mobileMaxInputLength
	case strings.Contains(lower, "otp"):
		return otpMaxInputLength
	default:
		return defaultMaxInputLength
	}
}

func isPasswordField(fieldPath string) bool {
	return strings.Contains(strings.ToLower(fieldPath), "password")
}

func joinFieldPath(parent string, field reflect.StructField) string {
	name := field.Name
	if tagValue, ok := parseJSONTag(field.Tag.Get("json")); ok {
		name = tagValue
	}
	if parent == "" {
		return name
	}
	return parent + "." + name
}

func parseJSONTag(tag string) (string, bool) {
	if tag == "" || tag == "-" {
		return "", false
	}
	comma := strings.Index(tag, ",")
	if comma == -1 {
		return tag, true
	}
	if comma == 0 {
		return "", false
	}
	return tag[:comma], true
}
