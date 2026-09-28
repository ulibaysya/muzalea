package storage

import (
	"github.com/pashagolub/pgxmock/v5"
	"slices"
	"testing"
)

func TestSave_Success(t *testing.T) {
	mock, err := pgxmock.NewPool(pgxmock.QueryMatcherOption(pgxmock.QueryMatcherEqual))
	if err != nil {
		t.Fatal(err)
	}
	defer mock.Close()

	name := "name"
	author := "author"
	tags := []string{
		"tag1",
		"tag2",
	}

	id := int64(5)

	mock.ExpectBegin()
	mock.ExpectQuery(`INSERT INTO audio (name, author) VALUES ($1, $2) RETURNING id;`).
		WithArgs(name, author).
		WillReturnRows(pgxmock.NewRows([]string{"id"}).AddRow(id))
	mock.ExpectExec(`INSERT INTO audio_tags (audio_id, tag) VALUES ($1, $2), ($3, $4);`).
		WithArgs(id, tags[0], id, tags[1]).
		WillReturnResult(pgxmock.NewResult("INSERT", 2))
	mock.ExpectCommit()

	ctx := t.Context()

	storage := New(mock)
	returnedID, returnedCompFunc, returnedErr := storage.Save(ctx, name, author, tags)

	if err = mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("asserting mock expectations: %v", err)
	}

	if returnedErr != nil {
		t.Fatalf("expected err nil, got %v", returnedErr)
	}

	if id != returnedID {
		t.Fatalf("expected id %v, got %v", id, returnedID)
	}

	if returnedCompFunc == nil {
		t.Fatalf("unexpected nil compensation func")
	}

	mock.ExpectExec(`DELETE FROM audio WHERE id = $1;`).
		WithArgs(id).
		WillReturnResult(pgxmock.NewResult("DELETE", 1))

	if returnCompErr := returnedCompFunc(ctx); returnCompErr != nil {
		t.Fatalf("expected compensation err nil, got %v", returnCompErr)
	}

	if err = mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("asserting mock expectations: %v", err)
	}
}

func TestSave_BuildTagsInsertSQL(t *testing.T) {
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
			name:    "SeveralTags",
			audioID: 5,
			tags:    []string{"tag1", "tag2"},
			sql:     `INSERT INTO audio_tags (audio_id, tag) VALUES ($1, $2), ($3, $4);`,
			args:    []any{int64(5), "tag1", int64(5), "tag2"},
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
