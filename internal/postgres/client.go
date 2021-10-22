package postgres

import (
	"database/sql"
)

type Options struct {
	ConnectionString string
	MaxIdleConns     int
	MaxOpenConns     int
}

func NewClient(opts Options) (*sql.DB, error) {
	db, err := sql.Open("postgres", opts.ConnectionString)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
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
