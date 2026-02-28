package domain

import "time"

var CurrentKeyID = "JWT_SECRET_V1"
var PrevKeyID = "JWT_SECRET_V0"

// BaseEntity defines model for BaseEntity
type BaseEntity struct {
	CreatedAt time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt time.Time  `json:"updated_at" db:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at" db:"deleted_at"`
} // @name BaseEntity
