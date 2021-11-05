package app

import (
	"context"

	"github.com/rislah/fakes/internal/errors"
)

func (r *roleImpl) CreateUserRole(ctx context.Context, userID string, roleID int) error {
	if userID == "" {
		return errors.New("userID is empty")
	}

	return r.roleDB.CreateUserRole(ctx, userID, roleID)
}
