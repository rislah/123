package app

import "context"

type UserDB interface {
	CreateUser(ctx context.Context, user User) error
	GetUsers(ctx context.Context) ([]User, error)
}
