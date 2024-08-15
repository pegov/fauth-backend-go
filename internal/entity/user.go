package entity

import "time"

type User struct {
	ID       int32   `db:"id"`
	Email    string  `db:"email"`
	Username string  `db:"username"`
	Password *string `db:"password"`

	// Roles
	// Permissions

	Active   bool `db:"active"`
	Verified bool `db:"verified"`

	CreatedAt time.Time `db:"created_at"`
	LastLogin time.Time `db:"last_login"`

	// OAuth
}
