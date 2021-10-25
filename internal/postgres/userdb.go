package postgres

import (
	"context"
	"database/sql"

	"github.com/cep21/circuit/v3"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	app "github.com/rislah/fakes/internal"
	"github.com/rislah/fakes/internal/errors"
)

type postgresUserDB struct {
	pg      *sqlx.DB
	circuit *circuit.Circuit
}

var _ app.UserDB = &postgresUserDB{}

func NewUserDB(pg *sqlx.DB, cc *circuit.Circuit) (*postgresUserDB, error) {
	pgUserDB := &postgresUserDB{pg: pg, circuit: cc}
	return pgUserDB, nil
}

func (p *postgresUserDB) CreateUser(ctx context.Context, user app.User) error {
	err := p.circuit.Run(ctx, func(c context.Context) error {
		tx, err := p.pg.BeginTx(ctx, &sql.TxOptions{})
		if err != nil {
			return err
		}

		res := tx.QueryRowContext(ctx, "insert into users (username, password_hash) VALUES ($1, $2) RETURNING user_id", user.Username, user.Password)
		if err != nil {
			return err
		}

		var userID string
		err = res.Scan(&userID)
		if err != nil {
			return err
		}

		_, err = tx.ExecContext(ctx, "insert into user_role (user_id) VALUES ($1)", userID)
		if err != nil {
			return err
		}

		err = tx.Commit()
		if err != nil {
			return err
		}

		return nil
	})
	return errors.New(err)
}

func (p *postgresUserDB) GetUsers(ctx context.Context) ([]app.User, error) {
	var users []app.User

	err := p.circuit.Run(ctx, func(c context.Context) error {
		err := p.pg.SelectContext(ctx, &users, `select u.user_id, u.username, u.password_hash from users u`)
		if err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return nil, errors.New(err)
	}

	return users, nil
}

func (p *postgresUserDB) GetUsersByIDs(ctx context.Context, ids []string) ([]app.User, error) {
	var users []app.User
	err := p.circuit.Run(ctx, func(c context.Context) error {
		query, args, err := sqlx.In(`select user_id, username, password_hash from users where user_id in (?);`, ids)
		if err != nil {
			return err
		}

		query = p.pg.Rebind(query)
		rows, err := p.pg.Query(query, args...)
		if err != nil {
			return err
		}

		for rows.Next() {
			user := app.User{}
			if err := rows.Scan(&user.UserID, &user.Username); err != nil {
				if err == sql.ErrNoRows {
					return nil
				}

				return err
			}

			users = append(users, user)
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return users, nil
}

func (p *postgresUserDB) GetUserRolesByUserIDs(ctx context.Context, userIDs []string) ([]app.Role, error) {
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

		query = p.pg.Rebind(query)
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
		return nil, err
	}

	return roles, nil
}

func (p *postgresUserDB) GetUserByUsername(ctx context.Context, username string) (app.User, error) {
	var user app.User
	err := p.circuit.Run(ctx, func(c context.Context) error {
		err := p.pg.GetContext(ctx, &user, `
			SELECT u.user_id, u.username, u.password_hash
			FROM users u
			WHERE username = $1
		`, username)
		if err != nil {
			if err == sql.ErrNoRows {
				return nil
			}
			return err
		}

		return nil
	})

	if err != nil {
		return app.User{}, errors.New(err)
	}

	return user, nil
}

func (p *postgresUserDB) GetUserRoleByUserID(ctx context.Context, userID string) (app.Role, error) {
	var role app.Role
	err := p.circuit.Run(ctx, func(c context.Context) error {
		err := p.pg.GetContext(ctx, &role, `
			select ur.user_id, r.id as role_id, r.name
			from user_role ur 
			inner join role r on r.id = ur.role_id
			where user_id = $1;
		`, userID)
		if err != nil {
			if err == sql.ErrNoRows {
				return nil
			}
			return err
		}

		return nil
	})

	if err != nil {
		return app.Role{}, errors.New(err)
	}

	return role, nil
}

func (p *postgresUserDB) GetUsersByRoleID(ctx context.Context, roleID int) ([]*app.User, error) {
	var users []*app.User
	err := p.circuit.Run(ctx, func(c context.Context) error {
		err := p.pg.SelectContext(ctx, &users, `
				select u.user_id, u.username, u.password_hash
				from users u
				inner join user_role ur on ur.user_id = u.user_id
				where ur.role_id = $1;
			`, roleID)
		if err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return users, nil
}

func (p *postgresUserDB) GetUsersByRoleIDs(ctx context.Context, roleIDs []int) ([]*app.UserRole, error) {
	var users []*app.UserRole
	err := p.circuit.Run(ctx, func(c context.Context) error {
		query, args, err := sqlx.In(`
			select u.user_id, u.username, u.password_hash, ur.role_id
			from users u
			inner join user_role ur on ur.user_id = u.user_id
			where ur.role_id in (?);
		`, roleIDs)
		if err != nil {
			return err
		}

		query = p.pg.Rebind(query)
		rows, err := p.pg.Query(query, args...)
		if err != nil {
			return err
		}

		for rows.Next() {
			user := app.UserRole{}
			if err := rows.Scan(&user.User.UserID, &user.User.Username, &user.User.Password, &user.Role.ID); err != nil {
				if err == sql.ErrNoRows {
					return nil
				}
				return err
			}
			users = append(users, &user)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return users, nil
}
