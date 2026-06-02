package ai

import (
	"strings"

	"booknest/internal/domain"
)

// BuildEmbeddingText builds the combined searchable text used for semantic embeddings.
//
// Format:
//   Title
//
//   Description
//
//   Summary
//
//   Category1 Category2 Category3
func BuildEmbeddingText(book domain.Book) string {
	title := trimOrFallback(book.Name, "Untitled book")
	description := trimOrFallback(book.Description, "No description available.")
	summary := trimOrFallback(book.Summary, "No summary available.")
	categories := joinCategoryNamesSpaceSeparated(book.Categories)

	var b strings.Builder
	b.WriteString(title)
	b.WriteString("\n\n")
	b.WriteString(description)
	b.WriteString("\n\n")
	b.WriteString(summary)
	b.WriteString("\n\n")
	b.WriteString(categories)
	return b.String()
}

func trimOrFallback(value, fallback string) string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return fallback
	}
	return trimmed
}

func joinCategoryNamesSpaceSeparated(categories []domain.Category) string {
	if len(categories) == 0 {
		return "No categories available."
	}

	names := make([]string, 0, len(categories))
	for _, category := range categories {
		name := strings.TrimSpace(category.Name)
		if name != "" {
			names = append(names, name)
		}
	}
	if len(names) == 0 {
		return "No categories available."
	}

	return strings.Join(names, " ")
}

