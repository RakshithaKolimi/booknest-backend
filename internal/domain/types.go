package domain

type SortOrder string

const (
	Asc  SortOrder = "asc"
	Desc SortOrder = "desc"
)

// SortOptions defines model for SortOptions
type SortOptions struct {
	Field string
	Order SortOrder
} // @name SortOptions

// QueryOptions defines model for QueryOptions
type QueryOptions struct {
	Limit  uint64
	Offset uint64
	Cursor *string
	Sort   *SortOptions
} // @name QueryOptions
