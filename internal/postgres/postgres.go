package postgres

import (
	"context"
	"database/sql"
	"github.com/pkg/errors"
	app "github.com/rislah/fakes/internal"
)

type Options struct {
	ConnectionString string
	MaxIdleConns     int
	MaxOpenConns     int
}

type postgresUserDB struct {
	pg *sql.DB
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

	return &postgresUserDB{pg: conn}, nil
}

func MakeUserDB(opts Options) (app.UserDB, error) {
	return NewUserDB(opts)
}

func (p *postgresUserDB) CreateUser(ctx context.Context, user app.User) error {
	_, err := p.pg.Exec("insert into users (firstname, lastname) VALUES ($1, $2)", user.Firstname, user.Lastname)
	if err != nil {
		return err
	}

	return err
}

func (p *postgresUserDB) GetUsers(ctx context.Context) ([]app.User, error) {
	var users []app.User

	rows, err := p.pg.Query("select firstname, lastname from users")
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		user := app.User{}
		if err := rows.Scan(&user.Firstname, &user.Lastname); err != nil {
			return nil, err
		}

		users = append(users, user)
	}

	return users, nil
}
