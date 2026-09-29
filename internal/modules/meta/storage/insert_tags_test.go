package storage

import (
	"errors"
	"log/slog"
	"slices"
	"testing"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/pashagolub/pgxmock/v5"
	"github.com/ulibaysya/muzalea/internal/modules/meta"
)

func TestInsertTags_DuplicateTags(t *testing.T) {
	mock, err := pgxmock.NewPool(pgxmock.QueryMatcherOption(pgxmock.QueryMatcherEqual))
	if err != nil {
		t.Fatal(err)
	}
	defer mock.Close()

	ctx := t.Context()
	mock.ExpectBegin()
	tx, err := mock.Begin(ctx)
	if err != nil {
		t.Fatal(err)
	}
	id := int64(123)
	tags := []string{
		"tag",
		"tag",
	}

	mock.ExpectExec(`INSERT INTO audio_tags (audio_id, tag) VALUES ($1, $2), ($3, $4);`).
		WithArgs(id, tags[0], id, tags[1]).
		WillReturnError(&pgconn.PgError{
			Severity:            "ERROR",
			SeverityUnlocalized: "ERROR",
			Code:                pgerrcode.UniqueViolation,
			Message:             "some message",
			ConstraintName:      `audio_tags_audio_id_tag_key`,
		})

	returnedErr := New(slog.Default(), mock).insertTags(ctx, tx, id, tags)

	if err = mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("asserting mock expectations: %v", err)
	}

	if !errors.Is(returnedErr, meta.ErrDuplicateTags) {
		t.Fatalf("expected err %v, got %v", meta.ErrDuplicateTags, returnedErr)
	}
}

func TestInsertTags_WrongAffectedNumber(t *testing.T) {
	mock, err := pgxmock.NewPool(pgxmock.QueryMatcherOption(pgxmock.QueryMatcherEqual))
	if err != nil {
		t.Fatal(err)
	}
	defer mock.Close()

	ctx := t.Context()
	mock.ExpectBegin()
	tx, err := mock.Begin(ctx)
	if err != nil {
		t.Fatal(err)
	}
	id := int64(123)
	tags := []string{
		"tag1",
		"tag2",
	}

	mock.ExpectExec(`INSERT INTO audio_tags (audio_id, tag) VALUES ($1, $2), ($3, $4);`).
		WithArgs(id, tags[0], id, tags[1]).
		WillReturnResult(pgxmock.NewResult("INSERT", 3))

	returnedErr := New(slog.Default(), mock).insertTags(ctx, tx, id, tags)

	if err = mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("asserting mock expectations: %v", err)
	}

	const expectedErrMsg = `unexpected number of rows is affected: expected 2, got 3`
	if errMsg := returnedErr.Error(); errMsg != expectedErrMsg {
		t.Fatalf("expected error message %v, got %v", expectedErrMsg, errMsg)
	}
}

func TestBuildTagsInsertSQL(t *testing.T) {
	tests := []struct {
		name    string
		audioID int64
		tags    []string
		sql     string
		args    []any
	}{
		{
			name: "NilTags",
			tags: nil,
			sql:  "",
			args: nil,
		},
		{
			name:    "OneTag",
			audioID: 5,
			tags:    []string{"tag1"},
			sql:     `INSERT INTO audio_tags (audio_id, tag) VALUES ($1, $2);`,
			args:    []any{int64(5), "tag1"},
		},
		{
			name:    "TwoTags",
			audioID: 5,
			tags:    []string{"tag1", "tag2"},
			sql:     `INSERT INTO audio_tags (audio_id, tag) VALUES ($1, $2), ($3, $4);`,
			args:    []any{int64(5), "tag1", int64(5), "tag2"},
		},
		{
			name:    "FiveTags",
			audioID: 5,
			tags: []string{
				"tag1",
				"tag2",
				"tag3",
				"tag4",
				"tag5",
			},
			sql:  `INSERT INTO audio_tags (audio_id, tag) VALUES ($1, $2), ($3, $4), ($5, $6), ($7, $8), ($9, $10);`,
			args: []any{int64(5), "tag1", int64(5), "tag2", int64(5), "tag3", int64(5), "tag4", int64(5), "tag5"},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			returnedSQL, returnedArgs := buildTagsInsertSQL(test.audioID, test.tags)

			if test.sql != returnedSQL {
				t.Fatalf("expected sql %v, got %v", test.sql, returnedSQL)
			}

			if !slices.Equal(test.args, returnedArgs) {
				t.Fatalf("args are not equal\n \t expected: %v\n \t got: %v", test.args, returnedArgs)
			}
		})
	}
}

func TestIsDuplicateTags(t *testing.T) {
	tests := []struct {
		name          string
		err           error
		wantDuplicate bool
	}{
		{
			name: "NotPgErr",
			err:  errors.New("not pg error"),
		},
		{
			name: "NotUniqueViolation",
			err: &pgconn.PgError{
				Code: pgerrcode.ForeignKeyViolation,
			},
		},
		{
			name: "InvalidConstraintName",
			err: &pgconn.PgError{
				Code:           pgerrcode.UniqueViolation,
				ConstraintName: "invalid constraint name",
			},
		},
		{
			name: "DuplicateError",
			err: &pgconn.PgError{
				Code:           pgerrcode.UniqueViolation,
				ConstraintName: "audio_tags_audio_id_tag_key",
			},
			wantDuplicate: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			isDuplicate := isDuplicateTags(test.err)
			if isDuplicate != test.wantDuplicate {
				t.Fatalf("expected %v, got %v", test.wantDuplicate, isDuplicate)
			}
		})
	}
}
