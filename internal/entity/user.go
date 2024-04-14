package entity

import "time"

type User struct {
	ID       int32
	Email    string
	Username string
	Password *string

	// Roles
	// Permissions

	Active   bool
	Verified bool

	CreatedAt time.Time
	LastLogin time.Time

	// OAuth
}
