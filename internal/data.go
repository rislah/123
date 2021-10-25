package app

import "github.com/jmoiron/sqlx"

type Data struct {
	Authenticator Authenticator
	UserDB        UserDB
	User          UserBackend
	DB            *sqlx.DB
}
