package storage

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/ulibaysya/muzalea/internal/pkg/compensate"
)

func (s Storage) Save(ctx context.Context, name, author string, tags []string) (id int64, cf compensate.Func, err error) {
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return 0, nil, fmt.Errorf("beginning tx: %v", err)
	}
	defer func() {
		if rollbackErr := tx.Rollback(ctx); rollbackErr != nil && !errors.Is(rollbackErr, pgx.ErrTxClosed) && err != nil {
			err = fmt.Errorf("rollbacking transaction: %v", rollbackErr)
		}
	}()

	const nameInsertSQL = `INSERT INTO audio (name, author) VALUES ($1, $2) RETURNING id;`
	var audioID int64
	if err = tx.QueryRow(ctx, nameInsertSQL, name, author).Scan(&audioID); err != nil {
		return 0, nil, fmt.Errorf("querying name and author insert: %v", err)
	}

	if tagsLen := len(tags); tagsLen != 0 {
		tagsInsertSQL, args := buildTagsInsertSQL(audioID, tags)
		res, err := tx.Exec(ctx, tagsInsertSQL, args...)
		if err != nil {
			return 0, nil, fmt.Errorf("inserting tags: %v", err)
		}
		if affected := res.RowsAffected(); affected != int64(tagsLen) {
			return 0, nil, fmt.Errorf("unexpected number of rows is affected when inserting tags: expected %v, got %v", tagsLen, affected)
		}
	}

	if err = tx.Commit(ctx); err != nil {
		return 0, nil, fmt.Errorf("commiting transaction: %v", err)
	}

	return audioID, saveCompensator{
		db:      s.db,
		audioID: audioID,
	}.compensate, nil
}

type saveCompensator struct {
	db      DB
	audioID int64
}

func (s saveCompensator) compensate(ctx context.Context) error {
	const sql = `DELETE FROM audio WHERE id = $1;`
	res, err := s.db.Exec(ctx, sql, s.audioID)
	if err != nil {
		return fmt.Errorf("deleting audio: %v", err)
	}

	if affected := res.RowsAffected(); affected != 1 {
		return fmt.Errorf("not 1 row affected when deleting: %v", affected)
	}

	return nil
}

func buildTagsInsertSQL(audioID int64, tags []string) (string, []any) {
	if len(tags) == 0 {
		return "", nil
	}

	b := &strings.Builder{}
	args := make([]any, 0, len(tags)*2)

	b.WriteString(`INSERT INTO audio_tags (audio_id, tag) VALUES ($1, $2)`)
	args = append(args, audioID, tags[0])

	const idArgOffset = 3
	const tagArgOffset = 4

	for i, tag := range tags[1:] {
		idArg := i + idArgOffset
		tagArg := i + tagArgOffset
		fmt.Fprintf(b, `, ($%d, $%d)`, idArg, tagArg)
		args = append(args, audioID, tag)
	}

	b.WriteString(";")

	return b.String(), args
}
