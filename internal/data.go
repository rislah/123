package app

type Data struct {
	Authenticator Authenticator
	User          UserBackend
	UserDB        UserDB
	RoleDB        RoleDB
}
