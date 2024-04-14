package model

type User struct {
	ID       int
	Username string
}

type UserCreate struct {
	Email    string
	Username string
	Password string
	Verified bool
}
