package app

import "github.com/jmoiron/sqlx"

type Data struct {
	Authenticator Authenticator
	UserDB        UserDB
	RoleDB        RoleDB
	User          UserBackend
	DB            *sqlx.DB
}
