package storage

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type DB interface {
	Exec(ctx context.Context, sql string, arguments ...any) (pgconn.CommandTag, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
	Begin(ctx context.Context) (pgx.Tx, error)
}

type Tx interface {
	DB
	Commit(ctx context.Context) error
}

type Storage struct {
	db DB
}

func New(db DB) Storage {
	return Storage{
		db: db,
	}
}
