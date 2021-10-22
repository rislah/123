package app

import "context"

type RBAC interface {
	CreateRole(ctx context.Context)
}
