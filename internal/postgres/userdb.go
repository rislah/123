package postgres

import (
	"context"
	"database/sql"

	"github.com/cep21/circuit/v3"
	"github.com/pkg/errors"
	app "github.com/rislah/fakes/internal"
)

type Options struct {
	ConnectionString string
	CircuitBreaker   *circuit.Circuit
	MaxIdleConns     int
	MaxOpenConns     int
}

type postgresUserDB struct {
	pg      *sql.DB
	circuit *circuit.Circuit
}

var _ app.UserDB = &postgresUserDB{}

func NewUserDB(opts Options) (*postgresUserDB, error) {
	if opts.ConnectionString == "" {
		return nil, errors.New("missing connection string")
	}

	conn, err := sql.Open("pq", opts.ConnectionString)
	if err != nil {
		return nil, err
	}

	conn.SetMaxIdleConns(opts.MaxIdleConns)
	conn.SetMaxOpenConns(opts.MaxOpenConns)

	if err := conn.Ping(); err != nil {
		return nil, err
	}

	return &postgresUserDB{pg: conn, circuit: opts.CircuitBreaker}, nil
}

func MakeUserDB(opts Options) (app.UserDB, error) {
	return NewUserDB(opts)
}

func (p *postgresUserDB) CreateUser(ctx context.Context, user app.User) error {
	err := p.circuit.Run(ctx, func(c context.Context) error {
		_, err := p.pg.Exec("insert into users (username, password) VALUES ($1, $2)", user.Username, user.Password)
		if err != nil {
			return err
		}
		return nil
	})
	return err
}

func (p *postgresUserDB) GetUsers(ctx context.Context) ([]app.User, error) {
	var (
		users []app.User
	)

	err := p.circuit.Run(ctx, func(c context.Context) error {
		rows, err := p.pg.Query("select u.username, p.password, r.name from users u inner join roles r on u.roles_id = r.id")
		if err != nil {
			return err
		}

		for rows.Next() {
			user := app.User{}
			if err := rows.Scan(&user.Username, &user.Password, &user.Role); err != nil {
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

func (p *postgresUserDB) GetUserByUsername(ctx context.Context, username string) (app.User, error) {
	var user app.User
	err := p.circuit.Run(ctx, func(c context.Context) error {
		row := p.pg.QueryRow("select username, password from users where username = $1", username)

		if err := row.Scan(&user.Username, &user.Password); err != nil {
			if err == sql.ErrNoRows {
				return &circuit.SimpleBadRequest{
					Err: err,
				}
			}
			return err
		}

		return nil
	})

	if err != nil {
		return app.User{}, err
	}

	return user, nil
}
