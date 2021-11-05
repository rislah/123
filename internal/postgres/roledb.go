package postgres

import (
	"context"
	"database/sql"

	"github.com/cep21/circuit"
	"github.com/jmoiron/sqlx"
	app "github.com/rislah/fakes/internal"
	"github.com/rislah/fakes/internal/errors"
	"github.com/rislah/fakes/internal/logger"
)

type roleDBImpl struct {
	pg      *sqlx.DB
	circuit *circuit.Circuit
}

func NewRoleDB(pg *sqlx.DB, cc *circuit.Circuit) *roleDBImpl {
	return &roleDBImpl{
		pg:      pg,
		circuit: cc,
	}
}

func (r *roleDBImpl) GetRoles(ctx context.Context) ([]app.Role, error) {
	var roles []app.Role
	err := r.circuit.Run(ctx, func(ctx context.Context) error {
		err := r.pg.SelectContext(ctx, &roles, `select id as role_id, name from role`)
		if err != nil {
			if err == sql.ErrNoRows {
				return nil
			}
			return err
		}

		return nil
	})

	if err != nil {
		return nil, errors.New(err)
	}

	return roles, nil
}

func (r *roleDBImpl) GetRolesByIDs(ctx context.Context, ids []int) ([]app.Role, error) {
	var roles []app.Role
	err := r.circuit.Run(ctx, func(ctx context.Context) error {
		query, args, err := sqlx.In(`
				select id as role_id, name
				from role
				where id in (?);
				`, ids)
		if err != nil {
			return err
		}

		query = r.pg.Rebind(query)
		rows, err := r.pg.Query(query, args...)
		if err != nil {
			return err
		}

		for rows.Next() {
			role := app.Role{}
			if err := rows.Scan(&role.ID, &role.Name); err != nil {
				if err == sql.ErrNoRows {
					return nil
				}

				return err
			}

			roles = append(roles, role)
		}

		return nil
	})

	if err != nil {
		return nil, errors.New(err)
	}

	return roles, nil
}

func (r *roleDBImpl) GetRolesByNames(ctx context.Context, names []string) ([]app.Role, error) {
	var roles []app.Role
	err := r.circuit.Run(ctx, func(c context.Context) error {
		query, args, err := sqlx.In(`
				select id as role_id, name
				from role
				where name in (?);
				`, names)
		if err != nil {
			return err
		}

		query = r.pg.Rebind(query)
		rows, err := r.pg.Query(query, args...)
		if err != nil {
			return err
		}

		for rows.Next() {
			role := app.Role{}
			if err := rows.Scan(&role.ID, &role.Name); err != nil {
				if err == sql.ErrNoRows {
					return nil
				}
				return err

			}

			roles = append(roles, role)
		}

		return nil
	})

	if err != nil {
		return nil, errors.New(err)
	}

	return roles, nil
}

func (r *roleDBImpl) GetRolesByUserIDs(ctx context.Context, userIDs []string) ([]app.Role, error) {
	var roles []app.Role
	err := r.circuit.Run(ctx, func(c context.Context) error {
		query, args, err := sqlx.In(`
			select ur.user_id, r.id, r.name
			from role r 
			inner join user_role ur on r.id = ur.role_id
			where user_id in (?);
		`, userIDs)
		if err != nil {
			return err
		}

		query = r.pg.Rebind(query)
		rows, err := r.pg.Query(query, args...)
		if err != nil {
			return err
		}

		for rows.Next() {
			role := app.Role{}
			if err := rows.Scan(&role.UserID, &role.ID, &role.Name); err != nil {
				if err == sql.ErrNoRows {
					return nil
				}

				return err
			}

			roles = append(roles, role)
		}

		return nil
	})

	if err != nil {
		return nil, errors.New(err)
	}

	return roles, nil
}

func (r *roleDBImpl) GetUserRoleByUserID(ctx context.Context, userID string) (app.Role, error) {
	var role app.Role
	err := r.circuit.Run(ctx, func(c context.Context) error {
		err := r.pg.GetContext(ctx, &role, `
			select ur.user_id, r.id as role_id, r.name
			from user_role ur 
			inner join role r on r.id = ur.role_id
			where user_id = $1;
		`, userID)
		if err != nil {
			if err == sql.ErrNoRows {
				return nil
			}
			logger.SharedGlobalLogger.Error("getuserorlebyuserid", err)
			return err
		}

		return nil
	})

	if err != nil {
		return app.Role{}, errors.New(err)
	}

	return role, nil
}

func (p *roleDBImpl) GetUserRolesByUserIDs(ctx context.Context, userIDs []string) ([]app.Role, error) {
	var roles []app.Role
	err := p.circuit.Run(ctx, func(c context.Context) error {
		query, args, err := sqlx.In(`
			select ur.user_id, r.id, r.name
			from user_role ur 
			inner join role r on r.id = ur.role_id
			where user_id in (?);
		`, userIDs)
		if err != nil {
			return err
		}

		rows, err := p.pg.Query(query, args...)
		if err != nil {
			return err
		}

		for rows.Next() {
			role := app.Role{}
			if err := rows.Scan(&role.UserID, &role.ID, &role.Name); err != nil {
				if err == sql.ErrNoRows {
					return nil
				}

				return err
			}

			roles = append(roles, role)
		}

		return nil
	})

	if err != nil {
		return nil, errors.New(err)
	}

	return roles, nil
}

func (r *roleDBImpl) CreateUserRole(ctx context.Context, userID string, roleID int) error {
	err := r.circuit.Run(ctx, func(c context.Context) error {
		tx, err := r.pg.BeginTx(ctx, &sql.TxOptions{})
		if err != nil {
			return err
		}

		_, err = tx.ExecContext(ctx, `insert into user_role (user_id, role_id) values ($1, $2)`, userID, roleID)
		if err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return errors.New(err)
	}

	return nil
}
