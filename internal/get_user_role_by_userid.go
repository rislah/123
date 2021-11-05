package app

import (
	"context"

	"github.com/rislah/fakes/internal/errors"
)

func (r *roleImpl) GetUserRoleByUserID(ctx context.Context, userID string) (Role, error) {
	if userID == "" {
		return Role{}, errors.New("userID is empty")
	}

	return r.roleDB.GetUserRoleByUserID(ctx, userID)
}
