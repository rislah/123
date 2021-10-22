package postgres

import (
	"context"
	"database/sql"

	"github.com/cep21/circuit/v3"
	_ "github.com/lib/pq"
	app "github.com/rislah/fakes/internal"
	"github.com/rislah/fakes/internal/errors"
)

type postgresUserDB struct {
	pg      *sql.DB
	circuit *circuit.Circuit
}

var _ app.UserDB = &postgresUserDB{}

func NewUserDB(pg *sql.DB, cc *circuit.Circuit) (*postgresUserDB, error) {
	pgUserDB := &postgresUserDB{pg: pg, circuit: cc}
	return pgUserDB, nil
}

func (p *postgresUserDB) CreateUser(ctx context.Context, user app.User) error {
	err := p.circuit.Run(ctx, func(c context.Context) error {
		_, err := p.pg.ExecContext(ctx, "insert into users (username, password_hash) VALUES ($1, $2)", user.Username, user.Password)
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
		rows, err := p.pg.QueryContext(ctx, `
			select u.user_id, u.username, u.password_hash, r.name
			from users u
			inner join roles r on u.role_id = r.id`)
		if err != nil {
			return err
		}

		for rows.Next() {
			user := app.User{}
			if err := rows.Scan(&user.UserID, &user.Username, &user.Password, &user.Role); err != nil {
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
		return nil, errors.New(err)
	}

	return users, nil
}

func (p *postgresUserDB) GetUserByUsername(ctx context.Context, username string) (app.User, error) {
	var user app.User
	err := p.circuit.Run(ctx, func(c context.Context) error {
		row := p.pg.QueryRowContext(ctx, `
            select u.user_id, u.username, u.password_hash, r.name 
            from users u
            inner join roles r on u.role_id = r.id
            where username = $1`, username)

		if err := row.Scan(&user.UserID, &user.Username, &user.Password, &user.Role); err != nil {
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
