package domain

import "time"

var CurrentKeyID = "JWT_SECRET_V1"
var PrevKeyID = "JWT_SECRET_V0"

var CurrentRefreshKeyID = "JWT_REFRESH_V1"
var PrevRefreshKeyID = "JWT_REFRESH_V0"

// BaseEntity defines model for BaseEntity
type BaseEntity struct {
	CreatedAt time.Time  `json:"created_at" db:"created_at" format:"date-time" example:"2026-04-06T10:30:00Z"`
	UpdatedAt time.Time  `json:"updated_at" db:"updated_at" format:"date-time" example:"2026-04-06T11:45:00Z"`
	DeletedAt *time.Time `json:"deleted_at" db:"deleted_at" format:"date-time" example:"2026-04-07T09:00:00Z"`
} // @name BaseEntity
