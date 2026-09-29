package storage

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/ulibaysya/muzalea/internal/modules/meta"
	"github.com/ulibaysya/muzalea/internal/pkg/compensate"
	pkgslog "github.com/ulibaysya/muzalea/pkg/slog"
)

func (s Storage) Save(ctx context.Context, name, author string, tags []string) (int64, compensate.Func, error) {
	tx, err := s.db.Begin(ctx)
	if err != nil {
		return 0, nil, fmt.Errorf("beginning tx: %w", err)
	}
	defer func() {
		if rollbackErr := tx.Rollback(ctx); rollbackErr != nil && !errors.Is(rollbackErr, pgx.ErrTxClosed) {
			s.logger.LogAttrs(ctx, slog.LevelError, "rollbacking transaction", pkgslog.Error("error", rollbackErr))
		}
	}()

	const nameInsertSQL = `INSERT INTO audio (name, author) VALUES ($1, $2) RETURNING id;`
	var audioID int64
	if err = tx.QueryRow(ctx, nameInsertSQL, name, author).Scan(&audioID); err != nil {
		return 0, nil, fmt.Errorf("inserting name and author: %w", err)
	}

	if len(tags) != 0 {
		if err = s.insertTags(ctx, tx, audioID, tags); err != nil {
			return 0, nil, fmt.Errorf("inserting tags: %w", err)
		}
	}

	if err = tx.Commit(ctx); err != nil {
		return 0, nil, fmt.Errorf("commiting transaction: %w", err)
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
		return fmt.Errorf("deleting audio: %w", err)
	}

	if affected := res.RowsAffected(); affected != 1 {
		return fmt.Errorf("not 1 row affected when deleting audio: %v", affected)
	}

	return nil
}

func (s Storage) insertTags(ctx context.Context, tx pgx.Tx, audioID int64, tags []string) error {
	tagsInsertSQL, args := buildTagsInsertSQL(audioID, tags)
	res, err := tx.Exec(ctx, tagsInsertSQL, args...)
	if err != nil {
		if isDuplicateTags(err) {
			return meta.ErrDuplicateTags
		} else {
			return err
		}
	}

	tagsLen := len(tags)
	if affected := res.RowsAffected(); affected != int64(tagsLen) {
		return fmt.Errorf("unexpected number of rows is affected: expected %v, got %v", tagsLen, affected)
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
		idArg := idArgOffset + i*2
		tagArg := tagArgOffset + i*2
		fmt.Fprintf(b, `, ($%d, $%d)`, idArg, tagArg)
		args = append(args, audioID, tag)
	}

	b.WriteString(";")

	return b.String(), args
}

func isDuplicateTags(err error) bool {
	pgErr, ok := errors.AsType[*pgconn.PgError](err)
	if !ok {
		return false
	}
	if pgErr.Code != pgerrcode.UniqueViolation {
		return false
	}
	const uniqueConstraintName = "audio_tags_audio_id_tag_key"
	return pgErr.ConstraintName == uniqueConstraintName
}
