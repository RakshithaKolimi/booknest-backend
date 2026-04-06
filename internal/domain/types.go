package domain

type SortOrder string

const (
	Asc  SortOrder = "asc"
	Desc SortOrder = "desc"
)

// SortOptions defines model for SortOptions
type SortOptions struct {
	Field string    `json:"field" example:"created_at"`
	Order SortOrder `json:"order" enums:"asc,desc" example:"desc"`
} // @name SortOptions

// QueryOptions defines model for QueryOptions
type QueryOptions struct {
	Limit  uint64       `json:"limit" example:"10"`
	Offset uint64       `json:"offset" example:"0"`
	Cursor *string      `json:"cursor,omitempty" example:"MjAyNi0wNC0wNlQxMTo0NTowMFo="`
	Sort   *SortOptions `json:"sort,omitempty"`
} // @name QueryOptions
