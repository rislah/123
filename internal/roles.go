package app

type Role string

func (r Role) String() string {
	return string(r)
}

const (
	DeveloperRole Role = "developer"
	UserRole      Role = "user"
	GuestRole     Role = "guest"
)
