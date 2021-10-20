package app

import "context"

type User struct {
	Username string `json:"username"`
	Password string `json:"password"`

	Firstname string
	Lastname  string
}

type UserDB interface {
	CreateUser(ctx context.Context, user User) error
	GetUsers(ctx context.Context) ([]User, error)
	GetUserByUsername(ctx context.Context, username string) (User, error)
}
