package app

type Backend struct {
	Authenticator Authenticator
	User          UserBackend
	Role          RoleBackend
}
