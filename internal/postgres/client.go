package postgres

import (
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

type Options struct {
	ConnectionString string
	MaxIdleConns     int
	MaxOpenConns     int
}

func NewClient(opts Options) (*sqlx.DB, error) {
	db, err := sqlx.Connect("postgres", opts.ConnectionString)
	if err != nil {
		return nil, err
	}

	if opts.MaxIdleConns == 0 {
		opts.MaxIdleConns = 100
	}

	if opts.MaxOpenConns == 0 {
		opts.MaxOpenConns = 100
	}

	db.SetMaxIdleConns(opts.MaxIdleConns)
	db.SetMaxOpenConns(opts.MaxOpenConns)

	return db, nil
}
